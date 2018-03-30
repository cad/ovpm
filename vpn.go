package ovpm

import (
	"bytes"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm/bindata"
	"github.com/cad/ovpm/pki"
	"github.com/cad/ovpm/supervisor"
	"github.com/coreos/go-iptables/iptables"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Possible VPN protocols.
const (
	TCPProto string = "tcp"
	UDPProto string = "udp"
)

// serverModel is database model for storing VPN server related stuff.
type dbServerModel struct {
	gorm.Model
	Name         string `gorm:"unique_index"` // Server name.
	SerialNumber string

	Hostname string // Server's ip address or FQDN
	Port     string // Server's listening port
	Proto    string // Server's proto udp or tcp
	Cert     string // Server RSA certificate.
	Key      string // Server RSA private key.
	CACert   string // Root CA RSA certificate.
	CAKey    string // Root CA RSA key.
	Net      string // VPN network.
	Mask     string // VPN network mask.
	CRL      string // Certificate Revocation List
	DNS      string // DNS servers to push to the clients.
}

// Server represents VPN server.
type Server struct {
	dbServerModel
	webPort string
}

// CheckSerial takes a serial number and checks it against the current server's serial number.
func (s *Server) CheckSerial(serial string) bool {
	return serial == s.SerialNumber
}

// GetSerialNumber returns server's serial number.
func (s *Server) GetSerialNumber() string {
	return s.SerialNumber
}

// GetServerName returns server's name.
func (s *Server) GetServerName() string {
	if s.Name != "" {
		return s.Name
	}
	return "default"
}

// GetHostname returns vpn server's hostname.
func (s *Server) GetHostname() string {
	return s.Hostname
}

// GetPort returns vpn server's port.
func (s *Server) GetPort() string {
	if s.Port != "" {
		return s.Port
	}
	return DefaultVPNPort

}

// GetProto returns vpn server's proto.
func (s *Server) GetProto() string {
	if s.Proto != "" {
		return s.Proto
	}
	return DefaultVPNProto
}

// GetCert returns vpn server's cert.
func (s *Server) GetCert() string {
	return s.Cert
}

// GetKey returns vpn server's key.
func (s *Server) GetKey() string {
	return s.Key
}

// GetCACert returns vpn server's cacert.
func (s *Server) GetCACert() string {
	return s.CACert
}

// GetCAKey returns vpn server's cakey.
func (s *Server) GetCAKey() string {
	return s.CAKey
}

// GetNet returns vpn server's net.
func (s *Server) GetNet() string {
	return s.Net
}

// GetMask returns vpn server's mask.
func (s *Server) GetMask() string {
	return s.Mask
}

// GetCRL returns vpn server's crl.
func (s *Server) GetCRL() string {
	return s.CRL
}

// GetDNS returns vpn server's dns.
func (s *Server) GetDNS() string {
	if s.DNS != "" {
		return s.DNS
	}
	return DefaultVPNDNS
}

// GetCreatedAt returns server's created at.
func (s *Server) GetCreatedAt() string {
	return s.CreatedAt.Format(time.UnixDate)
}

// Init regenerates keys and certs for a Root CA, gets initial settings for the VPN server
// and saves them in the database.
//
// 'proto' can be either "udp" or "tcp" and if it's "" it defaults to "udp".
//
// 'ipblock' is a IP network in the CIDR form. VPN clients get their IP addresses from this network.
// It defaults to const 'DefaultVPNNetwork'.
//
// Please note that, Init is potentially destructive procedure, it will cause invalidation of
// existing .ovpn profiles of the current users. So it should be used carefully.
func Init(hostname string, port string, proto string, ipblock string, dns string) error {
	if port == "" {
		port = DefaultVPNPort
	}

	if dns == "" {
		dns = DefaultVPNDNS
	}

	switch proto {
	case "":
		proto = UDPProto
	case UDPProto:
		proto = UDPProto
	case TCPProto:
		proto = TCPProto
	default:
		return fmt.Errorf("validation error: proto:`%s` should be either 'tcp' or 'udp'", proto)
	}

	// vpn network to use.
	var ipnet *net.IPNet

	// If user didn't specify, pick the vpn network from defaults.
	if ipblock == "" {
		var err error
		_, ipnet, err = net.ParseCIDR(DefaultVPNNetwork)
		if err != nil {
			return fmt.Errorf("can not parse CIDR %s: %v", DefaultVPNNetwork, err)
		}
	}

	// Check if the user specified vpn network is valid.
	if ipblock != "" && !govalidator.IsCIDR(ipblock) {
		return fmt.Errorf("validation error: ipblock:`%s` should be a CIDR network", ipblock)
	}

	// Use user specified vpn network.
	if ipblock != "" {
		var err error
		_, ipnet, err = net.ParseCIDR(ipblock)
		if err != nil {
			return fmt.Errorf("can parse ipblock: %s", err)

		}
	}

	if !govalidator.IsNumeric(port) {
		return fmt.Errorf("validation error: port:`%s` should be numeric", port)
	}

	serverName := "default"
	if IsInitialized() {
		if err := Deinit(); err != nil {
			logrus.Errorf("server can not be deleted: %v", err)
			return err
		}
	}

	if !govalidator.IsHost(hostname) {
		return fmt.Errorf("validation error: hostname:`%s` should be either an ip address or a FQDN", hostname)
	}

	if !govalidator.IsIPv4(dns) {
		return fmt.Errorf("validation error: dns:`%s` should be an ip address", dns)
	}

	ca, err := pki.NewCA()
	if err != nil {
		return fmt.Errorf("can not create ca creds: %s", err)
	}

	srv, err := pki.NewServerCertHolder(ca)
	if err != nil {
		return fmt.Errorf("can not create server cert creds: %s", err)
	}

	serialNumber := uuid.New().String()
	serverInstance := dbServerModel{
		Name: serverName,

		SerialNumber: serialNumber,
		Hostname:     hostname,
		Proto:        proto,
		Port:         port,
		Cert:         srv.Cert,
		Key:          srv.Key,
		CACert:       ca.Cert,
		CAKey:        ca.Key,
		Net:          ipnet.IP.To4().String(),
		Mask:         net.IP(ipnet.Mask).To4().String(),
		DNS:          dns,
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
		err := user.Renew()
		logrus.Infof("user certificate changed for %s, you should run: $ ovpm user export-config --user %s", user.Username, user.Username)
		if err != nil {
			logrus.Errorf("can not sign user %s: %v", user.Username, err)
			continue
		}
		// Set dynamic ip to user.
		user.HostID = 0
		db.Save(&user.dbUserModel)
	}
	EmitWithRestart()
	logrus.Infof("server initialized")
	return nil
}

// Update updates VPN server attributes.
func Update(ipblock string, dns string) error {
	if !IsInitialized() {
		return fmt.Errorf("server is not initialized")
	}

	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	var changed bool
	if ipblock != "" && govalidator.IsCIDR(ipblock) {
		var ipnet *net.IPNet
		_, ipnet, err = net.ParseCIDR(ipblock)
		if err != nil {
			return fmt.Errorf("can not parse CIDR %s: %v", ipblock, err)
		}
		server.dbServerModel.Net = ipnet.IP.To4().String()
		server.dbServerModel.Mask = net.IP(ipnet.Mask).To4().String()
		changed = true
	}

	if dns != "" && govalidator.IsIPv4(dns) {
		server.dbServerModel.DNS = dns
		changed = true
	}
	if changed {
		db.Save(server.dbServerModel)
		users, err := GetAllUsers()
		if err != nil {
			return err
		}

		// Set all users to dynamic ip address.
		// This way we prevent any ip range mismatch.
		for _, user := range users {
			user.HostID = 0
			db.Save(user.dbUserModel)
		}

		EmitWithRestart()
		logrus.Infof("server updated")
	}
	return nil
}

// Deinit deletes the VPN server from the database and frees the allocated resources.
func Deinit() error {
	if !IsInitialized() {
		return fmt.Errorf("server not found")
	}

	db.Unscoped().Delete(&dbServerModel{})
	db.Unscoped().Delete(&dbRevokedModel{})
	EmitWithRestart()
	return nil
}

// DumpsClientConfig generates .ovpn file for the given vpn user and returns it as a string.
func DumpsClientConfig(username string) (string, error) {
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
		NoGW     bool
		Proto    string
	}{
		Hostname: server.GetHostname(),
		Port:     server.GetPort(),
		CA:       server.GetCACert(),
		Key:      user.getKey(),
		Cert:     user.GetCert(),
		NoGW:     user.IsNoGW(),
		Proto:    server.GetProto(),
	}
	data, err := bindata.Asset("template/client.ovpn.tmpl")
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

// DumpClientConfig generates .ovpn file for the given vpn user and dumps it to outPath.
func DumpClientConfig(username, path string) error {
	result, err := DumpsClientConfig(username)
	if err != nil {
		return err
	}
	// Wite rendered content into openvpn server conf.
	return emitToFile(path, result, 0)

}

// GetSystemCA returns the system CA from the database if available.
func GetSystemCA() (*pki.CA, error) {
	server := dbServerModel{}
	db.First(&server)
	if db.NewRecord(&server) {
		return nil, fmt.Errorf("server record does not exists in db")
	}
	return &pki.CA{
		CertHolder: pki.CertHolder{
			Cert: server.CACert,
			Key:  server.CAKey,
		},
	}, nil

}

// vpnProc represents the OpenVPN process that is managed by the ovpm supervisor globally OpenVPN.
var vpnProc supervisor.Supervisable

// StartVPNProc starts the OpenVPN process.
func StartVPNProc() {
	if !IsInitialized() {
		logrus.Error("can not launch OpenVPN because system is not initialized")
		return
	}
	if vpnProc == nil {
		panic(fmt.Sprintf("vpnProc is not initialized!"))
	}
	if vpnProc.Status() == supervisor.RUNNING {
		logrus.Error("OpenVPN is already started")
		return
	}
	Emit()
	vpnProc.Start()
	ensureNatEnabled()
}

// RestartVPNProc restarts the OpenVPN process.
func RestartVPNProc() {
	if !IsInitialized() {
		logrus.Error("can not launch OpenVPN because system is not initialized")
		return
	}
	if vpnProc == nil {
		panic(fmt.Sprintf("vpnProc is not initialized!"))
	}
	Emit()
	vpnProc.Restart()
	ensureNatEnabled()
}

// StopVPNProc stops the OpenVPN process.
func StopVPNProc() {
	if vpnProc == nil {
		panic(fmt.Sprintf("vpnProc is not initialized!"))
	}
	if vpnProc.Status() != supervisor.RUNNING {
		logrus.Error("OpenVPN is already not running")
		return
	}
	vpnProc.Stop()
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

	if !IsInitialized() {
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
		return fmt.Errorf("can not emit iptables: %s", err)
	}

	if err := emitCRL(); err != nil {
		return fmt.Errorf("can not emit crl: %s", err)
	}

	logrus.Info("configurations emitted to the filesystem")
	return nil
}

// EmitWithRestart restarts vpnProc after calling EmitWithRestart().
func EmitWithRestart() error {
	err := Emit()
	if err != nil {
		return err
	}
	if IsInitialized() {
		for {
			if vpnProc.Status() == supervisor.RUNNING || vpnProc.Status() == supervisor.STOPPED {
				logrus.Info("OpenVPN process is restarting")
				RestartVPNProc()
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	return nil

}

func emitToFile(path, content string, mode uint) error {
	// When testing don't emit files to the filesystem. Just pretend you did.
	if Testing {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Cannot create file %s: %v", path, err)

	}
	if mode != 0 {
		file.Chmod(os.FileMode(mode))
	}
	defer file.Close()
	fmt.Fprintf(file, content)
	return nil
}

func emitServerConf() error {
	dbServer, err := GetServerInstance()
	if err != nil {
		return fmt.Errorf("can not get server instance: %v", err)
	}

	serverInstance, err := GetServerInstance()
	if err != nil {
		return fmt.Errorf("can not retrieve server: %v", err)
	}
	port := DefaultVPNPort
	if serverInstance.Port != "" {
		port = serverInstance.Port
	}

	proto := DefaultVPNProto
	if serverInstance.Proto != "" {
		proto = serverInstance.Proto
	}

	dns := DefaultVPNDNS
	if serverInstance.DNS != "" {
		dns = serverInstance.DNS
	}

	var result bytes.Buffer

	server := struct {
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
		Proto        string
		DNS          string
	}{
		CertPath:     _DefaultCertPath,
		KeyPath:      _DefaultKeyPath,
		CACertPath:   _DefaultCACertPath,
		CAKeyPath:    _DefaultCAKeyPath,
		CCDPath:      _DefaultVPNCCDPath,
		CRLPath:      _DefaultCRLPath,
		DHParamsPath: _DefaultDHParamsPath,
		Net:          dbServer.Net,
		Mask:         dbServer.Mask,
		Port:         port,
		Proto:        proto,
		DNS:          dns,
	}
	data, err := bindata.Asset("template/server.conf.tmpl")
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
	return emitToFile(_DefaultVPNConfPath, result.String(), 0)
}

// GetServerInstance returns the default server from the database.
func GetServerInstance() (*Server, error) {
	var server dbServerModel
	db.First(&server)
	if db.NewRecord(server) {
		return nil, fmt.Errorf("can not retrieve server from db")
	}
	return &Server{dbServerModel: server}, nil
}

// GetConnectedUsers will return a list of users who are currently connected
// to the VPN service.
func GetConnectedUsers() ([]User, error) {
	var users []User

	// Open the status log file.
	f, err := os.Open(_DefaultStatusLogPath)
	if err != nil {
		panic(err)
	}

	cl, _ := parseStatusLog(f)
	for _, c := range cl {
		var u dbUserModel
		q := db.Where(dbUserModel{Username: c.CommonName}).First(&u)
		if q.RecordNotFound() {
			return nil, fmt.Errorf("common name reported by the OpenVPN is not found in the database: username='%s'", c.CommonName)
		}
		if err := q.Error; err != nil {
			return nil, fmt.Errorf("unknown db error: %v", err)
		}

		users = append(users, User{
			dbUserModel:    u,
			isConnected:    true,
			connectedSince: c.ConnectedSince,
			bytesReceived:  c.BytesReceived,
			bytesSent:      c.BytesSent,
		})
	}
	return users, nil
}

// IsInitialized checks if there is a default VPN server configured in the database or not.
func IsInitialized() bool {
	var server dbServerModel
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
	return emitToFile(_DefaultKeyPath, server.Key, 0600)
}

func emitServerCert() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the cert file.
	return emitToFile(_DefaultCertPath, server.Cert, 0)
}

func emitCRL() error {
	var revokedDBItems []*dbRevokedModel
	db.Find(&revokedDBItems)
	var revokedCertSerials []*big.Int
	for _, item := range revokedDBItems {
		bi := big.NewInt(0)
		bi.SetString(item.SerialNumber, 16)
		revokedCertSerials = append(revokedCertSerials, bi)
	}
	systemCA, err := GetSystemCA()
	if err != nil {
		return fmt.Errorf("can not emit CRL: %v", err)
	}
	crl, err := pki.NewCRL(systemCA, revokedCertSerials...)
	if err != nil {
		return fmt.Errorf("can not emit crl: %v", err)
	}

	return emitToFile(_DefaultCRLPath, crl, 0)
}

func emitCACert() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the ca cert file.
	return emitToFile(_DefaultCACertPath, server.CACert, 0)
}

func emitCAKey() error {
	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	// Write rendered content into the ca key file.
	return emitToFile(_DefaultCAKeyPath, server.CAKey, 0600)
}

func emitCCD() error {
	users, err := GetAllUsers()
	if err != nil {
		return err
	}

	// Filesystem related stuff. Skipping when testing.
	if !Testing {
		// Clean and then create and write rendered ccd data.
		err = os.RemoveAll(_DefaultVPNCCDPath)
		if err != nil {
			if os.IsNotExist(err) {
			} else {
				return err
			}
		}

		if _, err := os.Stat(_DefaultVPNCCDPath); err != nil {
		}

		err = os.Mkdir(_DefaultVPNCCDPath, 0755)
		if err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}
	server, err := GetServerInstance()
	if err != nil {
		return fmt.Errorf("can not get server instance: %v", err)
	}

	// Render ccd templates for the users.
	for _, user := range users {
		var associatedRoutes [][3]string
		var serverNets [][2]string
		for _, network := range GetAllNetworks() {
			switch network.Type {
			case ROUTE:
				for _, assocUsername := range network.GetAssociatedUsernames() {
					if assocUsername == user.Username {
						via := network.Via
						ip, mask, err := net.ParseCIDR(network.CIDR)
						if err != nil {
							return err
						}
						associatedRoutes = append(associatedRoutes, [3]string{ip.To4().String(), net.IP(mask.Mask).To4().String(), via})
					}
				}
			case SERVERNET:
				// Push associated servernets to client when client is not getting vpn server as default gw.
				if user.IsNoGW() {
					for _, assocUsername := range network.GetAssociatedUsernames() {
						if assocUsername == user.Username {
							ip, mask, err := net.ParseCIDR(network.CIDR)
							if err != nil {
								return err
							}
							serverNets = append(serverNets, [2]string{ip.To4().String(), net.IP(mask.Mask).To4().String()})
						}
					}
				}
			}
		}
		var result bytes.Buffer
		params := struct {
			IP         string
			NetMask    string
			Routes     [][3]string // [0] is IP, [1] is Netmask, [2] is Via
			Servernets [][2]string // [0] is IP, [1] is Netmask
			RedirectGW bool
		}{IP: user.getIP().String(), NetMask: server.Mask, Routes: associatedRoutes, Servernets: serverNets, RedirectGW: !user.NoGW}

		data, err := bindata.Asset("template/ccd.file.tmpl")
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

		err = emitToFile(filepath.Join(_DefaultVPNCCDPath, user.Username), result.String(), 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func emitDHParams() error {
	var result bytes.Buffer
	data, err := bindata.Asset("template/dh4096.pem.tmpl")
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

	err = emitToFile(_DefaultDHParamsPath, result.String(), 0)
	if err != nil {
		return err
	}
	return nil
}

func emitIptables() error {
	if Testing {
		return nil
	}
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("can not create new iptables object: %v", err)
	}

	for _, network := range GetAllNetworks() {
		associatedUsernames := network.GetAssociatedUsernames()
		switch network.Type {
		case SERVERNET:
			users, err := GetAllUsers()
			if err != nil {
				return err
			}

			// Find associated users and emit iptables configs for the users
			// regarding the network's type and attributes.
			for _, user := range users {
				// Find out if the user is associated or not.
				var found bool
				for _, auser := range associatedUsernames {
					if user.Username == auser {
						found = true
						break
					}
				}

				userIP, _, err := net.ParseCIDR(user.GetIPNet())
				if err != nil {
					return err
				}
				_, networkIPNet, err := net.ParseCIDR(network.CIDR)
				if err != nil {
					return err
				}

				// get destination network's iface
				iface := interfaceOfIP(networkIPNet)
				if iface == nil {
					logrus.Warnf("network doesn't exist on server %s[SERVERNET]: cant find interface for %s", network.Name, networkIPNet.String())
					return nil
				}
				// enable nat for the user to the destination network n
				if found {
					err = ipt.AppendUnique("nat", "POSTROUTING", "-s", userIP.String(), "-o", iface.Name, "-j", "MASQUERADE")
					if err != nil {
						logrus.Error(err)
						return err
					}
				} else {
					err = ipt.Delete("nat", "POSTROUTING", "-s", userIP.String(), "-o", iface.Name, "-j", "MASQUERADE")
					if err != nil {
						logrus.Debug(err)
					}
				}
			}
		}
	}
	return nil
}

func checkOpenVPNExecutable() bool {
	executable := getOpenVPNExecutable()
	if executable == "" {
		logrus.Error("openvpn is not installed ✘")
		return false
	}
	logrus.Debugf("openvpn executable detected: %s  ✔", executable)
	return true
}

func getOpenVPNExecutable() string {
	cmd := exec.Command("which", "openvpn")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("openvpn is not installed: %s  ✘", err)
		return ""
	}
	logrus.Debugf("openvpn executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return strings.TrimSpace(string(output[:]))
}

func checkOpenSSLExecutable() bool {
	cmd := exec.Command("which", "openssl")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("openssl is not installed: %s  ✘", err)
		return false
	}
	logrus.Debugf("openssl executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return true
}

func checkIptablesExecutable() bool {
	cmd := exec.Command("which", "iptables")
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables is not installed: %s  ✘", err)
		return false
	}
	logrus.Debugf("iptables executable detected: %s  ✔", strings.TrimSpace(string(output[:])))
	return true
}

func ensureBaseDir() {
	if Testing {
		return
	}
	os.Mkdir(varBasePath, 0755)
}

func init() {
	ensureBaseDir()
	var err error
	vpnProc, err = supervisor.NewProcess(getOpenVPNExecutable(), varBasePath, []string{"--config", _DefaultVPNConfPath})
	if err != nil {
		logrus.Errorf("can not create process: %v", err)
	}
}
