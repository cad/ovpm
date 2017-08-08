//go:generate go-bindata -pkg ovpm template/

package ovpm

import (
	"bytes"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// DBNetwork is database model for external networks on the VPN server.
type DBNetwork struct {
	gorm.Model
	ServerID uint
	Server   DBServer

	Name        string
	NetworkCIDR string
}

// DBServer is database model for storing VPN server related stuff.
type DBServer struct {
	gorm.Model
	Name         string `gorm:"unique_index"` // Server name.
	SerialNumber string

	Hostname string // Server's ip address or FQDN
	Port     string // Server's listening port
	Cert     string // Server RSA certificate.
	Key      string // Server RSA private key.
	CACert   string // Root CA RSA certificate.
	CAKey    string // Root CA RSA key.
	Net      string // VPN network.
	Mask     string // VPN network mask.
	CRL      string // Certificate Revocation List
}

// CheckSerial takes a serial number and checks it against the current server's serial number.
func (s *DBServer) CheckSerial(serialNo string) bool {
	return serialNo == s.SerialNumber
}

type _VPNServerConfig struct {
	CertPath     string
	KeyPath      string
	CACertPath   string
	CAKeyPath    string
	CCDPath      string
	CRLPath      string
	DHParamsPath string
	Net          string
	Mask         string
	Port         string
}

// InitServer regenerates keys and certs for a Root CA, and saves them in the database.
func InitServer(serverName string, hostname string, port string) error {
	if CheckBootstrapped() {
		if err := DeleteServer("default"); err != nil {
			logrus.Errorf("server can not be deleted: %v", err)
			return err
		}
	}

	if !govalidator.IsHost(hostname) {
		return fmt.Errorf("validation error: hostname:`%s` should be either an ip address or a FQDN", hostname)
	}

	ca, err := CreateCA()
	if err != nil {
		return fmt.Errorf("can not create ca creds: %s", err)
	}

	srv, err := CreateServerCert(ca)
	if err != nil {
		return fmt.Errorf("can not create server cert creds: %s", err)
	}
	serialNumber := uuid.New().String()

	// crl, err := MakeCRL([]*big.Int{})
	// if err != nil {
	// 	return fmt.Errorf("can not create server crl: %v", err)
	// }

	serverInstance := DBServer{
		Name: serverName,

		SerialNumber: serialNumber,
		Hostname:     hostname,
		Port:         port,
		Cert:         srv.Cert,
		Key:          srv.Key,
		CACert:       ca.Cert,
		CAKey:        ca.Key,
		Net:          DefaultServerNetwork,
		Mask:         DefaultServerNetMask,
	}

	db.Create(&serverInstance)

	if db.NewRecord(&serverInstance) {
		return fmt.Errorf("can not create server instance on database")
	}

	users, err := GetAllUsers()
	if err != nil {
		return err
	}
	// Sign all users in the db with the new server
	for _, user := range users {
		err := user.Sign()
		logrus.Infof("user certificate changed for %s, you should run: $ ovpm user export-config --user %s", user.Username, user.Username)
		if err != nil {
			logrus.Errorf("can not sign user %s: %v", user.Username, err)
			continue
		}
	}
	return nil
}

// DeleteServer deletes the server with the given serverName from the database.
func DeleteServer(serverName string) error {
	if !CheckBootstrapped() {
		return fmt.Errorf("server not found")
	}

	db.Unscoped().Delete(&DBServer{})
	db.Unscoped().Delete(&DBRevoked{})
	return nil
}

func sDumpUserOVPNConf(username string) (string, error) {
	var result bytes.Buffer
	user, err := GetUser(username)
	if err != nil {
		return "", err
	}

	server, err := GetServerInstance()
	if err != nil {
		return "", err
	}

	params := struct {
		Hostname string
		Port     string
		CA       string
		Key      string
		Cert     string
	}{
		Hostname: server.Hostname,
		Port:     server.Port,
		CA:       server.CACert,
		Key:      user.Key,
		Cert:     user.Cert,
	}
	data, err := Asset("template/client.ovpn.tmpl")
	if err != nil {
		return "", err
	}

	t, err := template.New("client.ovpn").Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("can not parse client.ovpn.tmpl template: %s", err)
	}

	err = t.Execute(&result, params)
	if err != nil {
		return "", fmt.Errorf("can not render client.ovpn: %s", err)
	}

	return result.String(), nil
}

// DumpUserOVPNConf combines a specially generated config for the client with CA's and Client's certs and Clients key then dumps them to the specified path.
func DumpUserOVPNConf(username, outPath string) error {
	result, err := sDumpUserOVPNConf(username)
	if err != nil {
		return err
	}
	// Wite rendered content into openvpn server conf.
	return emitToFile(outPath, result, 0)

}

// Emit generates all needed files for the OpenVPN server and dumps them to their corresponding paths defined in the config.
func Emit() error {
	// Check dependencies
	if !checkOpenVPNExecutable() {
		return fmt.Errorf("openvpn executable can not be found! you should install OpenVPN on this machine")
	}

	if !checkOpenSSLExecutable() {
		return fmt.Errorf("openssl executable can not be found! you should install openssl on this machine")

	}

	if !checkIptablesExecutable() {
		return fmt.Errorf("iptables executable can not be found")
	}

	if !CheckBootstrapped() {
		return fmt.Errorf("you should create a server first. e.g. $ ovpm vpn create-server")
	}

	if err := emitServerConf(); err != nil {
		return fmt.Errorf("can not emit server conf: %s", err)
	}

	if err := emitServerCert(); err != nil {
		return fmt.Errorf("can not emit server cert: %s", err)
	}

	if err := emitServerKey(); err != nil {
		return fmt.Errorf("can not emit server key: %s", err)
	}

	if err := emitCACert(); err != nil {
		return fmt.Errorf("can not emit ca cert : %s", err)
	}

	if err := emitCAKey(); err != nil {
		return fmt.Errorf("can not emit ca key: %s", err)
	}

	if err := emitDHParams(); err != nil {
		return fmt.Errorf("can not emit dhparams: %s", err)
	}

	if err := emitCCD(); err != nil {
		return fmt.Errorf("can not emit ccd: %s", err)
	}

	if err := emitIptables(); err != nil {
		return fmt.Errorf("can not emit iptables conf: %s", err)
	}

	if err := emitCRL(); err != nil {
		return fmt.Errorf("can not emit crl: %s", err)
	}

	logrus.Info("changes are applied to the filesystem")

	return nil
}

func emitToFile(filePath, content string, mode uint) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Cannot create file %s: %v", filePath, err)

	}
	if mode != 0 {
		file.Chmod(os.FileMode(mode))
	}
	defer file.Close()
	fmt.Fprintf(file, content)
	return nil
}

func emitServerConf() error {
	serverInstance, err := GetServerInstance()
	if err != nil {
		return fmt.Errorf("can not retrieve server: %v", err)
	}
	port := DefaultVPNPort
	if serverInstance.Port != "" {
		port = serverInstance.Port
	}

	var result bytes.Buffer

	server := _VPNServerConfig{
		CertPath:     DefaultCertPath,
		KeyPath:      DefaultKeyPath,
		CACertPath:   DefaultCACertPath,
		CAKeyPath:    DefaultCAKeyPath,
		CCDPath:      DefaultVPNCCDPath,
		CRLPath:      DefaultCRLPath,
		DHParamsPath: DefaultDHParamsPath,
		Net:          DefaultServerNetwork,
		Mask:         DefaultServerNetMask,
		Port:         port,
	}
	data, err := Asset("template/server.conf.tmpl")
	if err != nil {
		return err
	}

	t, err := template.New("server.conf").Parse(string(data))
	if err != nil {
		return fmt.Errorf("can not parse server.conf.tmpl template: %s", err)
	}

	err = t.Execute(&result, server)
	if err != nil {
		return fmt.Errorf("can not render server.conf: %s", err)
	}

	// Wite rendered content into openvpn server conf.
	return emitToFile(DefaultVPNConfPath, result.String(), 0)
}

// GetServerInstance returns the default server from the database.
func GetServerInstance() (*DBServer, error) {
	var server DBServer
	db.First(&server)
	if db.NewRecord(server) {
		return nil, fmt.Errorf("can not retrieve server from db")
	}
	return &server, nil
}

// CheckBootstrapped checks if there is a default server in the database or not.
func CheckBootstrapped() bool {
	var server DBServer
	db.First(&server)
	if db.NewRecord(server) {
		return false
	}
	return true
}

func emitServerKey() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into key file.
	return emitToFile(DefaultKeyPath, server.Key, 0600)
}

func emitServerCert() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the cert file.
	return emitToFile(DefaultCertPath, server.Cert, 0)
}

func emitCRL() error {
	var revokedDBItems []*DBRevoked
	db.Find(&revokedDBItems)
	var revokedCertSerials []*big.Int
	for _, item := range revokedDBItems {
		bi := big.NewInt(0)
		bi.SetString(item.SerialNumber, 16)
		revokedCertSerials = append(revokedCertSerials, bi)
	}
	crl, err := MakeCRL(revokedCertSerials)
	if err != nil {
		return fmt.Errorf("can not emit crl: %v", err)
	}

	return emitToFile(DefaultCRLPath, crl, 0)
}

func emitCACert() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the ca cert file.
	return emitToFile(DefaultCACertPath, server.CACert, 0)
}

func emitCAKey() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the ca key file.
	return emitToFile(DefaultCAKeyPath, server.CAKey, 0600)
}

func emitCCD() error {
	users, err := GetAllUsers()
	if err != nil {
		return err
	}

	// Create and write rendered ccd data.
	os.Mkdir(DefaultVPNCCDPath, 0755)
	clientsNetMask := net.IPMask(net.ParseIP(DefaultServerNetMask))
	clientsNetPrefix := net.ParseIP(DefaultServerNetwork)
	clientNet := clientsNetPrefix.Mask(clientsNetMask).To4()

	counter := 2
	for _, user := range users {
		var result bytes.Buffer
		clientNet[3] = byte(counter)
		params := struct {
			IP      string
			NetMask string
		}{IP: clientNet.String(), NetMask: DefaultServerNetMask}

		data, err := Asset("template/ccd.file.tmpl")
		if err != nil {
			return err
		}
		t, err := template.New("ccd.file.tmpl").Parse(string(data))
		if err != nil {
			return fmt.Errorf("can not parse ccd.file.tmpl template: %s", err)
		}

		err = t.Execute(&result, params)
		if err != nil {
			return fmt.Errorf("can not render ccd file %s: %s", user.Username, err)
		}

		err = emitToFile(DefaultVPNCCDPath+user.Username, result.String(), 0)
		if err != nil {
			return err
		}
		counter++
	}
	return nil
}

func emitDHParams() error {
	var result bytes.Buffer
	data, err := Asset("template/dh4096.pem.tmpl")
	if err != nil {
		return err
	}

	t, err := template.New("dh4096.pem.tmpl").Parse(string(data))
	if err != nil {
		return fmt.Errorf("can not parse dh4096.pem template: %s", err)
	}

	err = t.Execute(&result, nil)
	if err != nil {
		return fmt.Errorf("can not render dh4096.pem file: %s", err)
	}

	err = emitToFile(DefaultDHParamsPath, result.String(), 0)
	if err != nil {
		return err
	}
	return nil
}

func emitIptables() error {
	return nil
}

func checkOpenVPNExecutable() bool {
	cmd := exec.Command("which", "openvpn")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("openvpn is not installed: %s  ✘", err)
		return false
	}
	logrus.Infof("openvpn executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return true
}

func checkOpenSSLExecutable() bool {
	cmd := exec.Command("which", "openssl")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("openssl is not installed: %s  ✘", err)
		return false
	}
	logrus.Infof("openssl executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return true
}

func checkIptablesExecutable() bool {
	cmd := exec.Command("which", "iptables")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables is not installed: %s  ✘", err)
		return false
	}
	logrus.Infof("iptables executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return true
}
