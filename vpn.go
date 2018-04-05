package ovpm

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

var serverInstance *Server
var once sync.Once

// Server represents VPN server.
type Server struct {
	dbServerModel

	webPort string

	emitToFileFunc     func(path, content string, mode uint) error
	openFunc           func(path string) (io.Reader, error)
	parseStatusLogFunc func(f io.Reader) ([]clEntry, []rtEntry)
}

// TheServer returns a pointer to the server instance.
//
// Server instance is a singleton instance that is initialized
// on the first call made to the TheServer().
func TheServer() *Server {
	once.Do(func() {
		// Initialize the server instance by setting default mockable funcs & attributes.
		serverInstance = &Server{
			emitToFileFunc: emitToFile,
			openFunc: func(path string) (io.Reader, error) {
				return os.Open(path)
			},
			parseStatusLogFunc: parseStatusLog,
		}
	})
	if db != nil {
		serverInstance.Refresh()
	} else {
		logrus.Warn("database is not connected yet. skipping server instance refresh")
	}
	return serverInstance
}

// CheckSerial takes a serial number and checks it against the current server's serial number.
func (svr *Server) CheckSerial(serial string) bool {
	return serial == svr.SerialNumber
}

// GetSerialNumber returns server's serial number.
func (svr *Server) GetSerialNumber() string {
	return svr.SerialNumber
}

// GetServerName returns server's name.
func (svr *Server) GetServerName() string {
	if svr.Name != "" {
		return svr.Name
	}
	return "default"
}

// GetHostname returns vpn server's hostname.
func (svr *Server) GetHostname() string {
	return svr.Hostname
}

// GetPort returns vpn server's port.
func (svr *Server) GetPort() string {
	if svr.Port != "" {
		return svr.Port
	}
	return DefaultVPNPort

}

// GetProto returns vpn server's proto.
func (svr *Server) GetProto() string {
	if svr.Proto != "" {
		return svr.Proto
	}
	return DefaultVPNProto
}

// CAExpiresAt returns the expiry date time of the CA.
func (svr *Server) CAExpiresAt() time.Time {
	if !svr.IsInitialized() {
		return time.Time{}
	}
	crt, err := pki.ReadCertFromPEM(svr.CACert)
	if err != nil {
		logrus.Fatalf("can't parse cert: %v", err)
	}
	return crt.NotAfter
}

// ExpiresAt returns the expiry date time of the server cert.
func (svr *Server) ExpiresAt() time.Time {
	if !svr.IsInitialized() {
		return time.Time{}
	}
	crt, err := pki.ReadCertFromPEM(svr.Cert)
	if err != nil {
		logrus.Fatalf("can't parse cert: %v", err)
	}
	return crt.NotAfter
}

// GetKey returns vpn server's key.
func (svr *Server) GetKey() string {
	return svr.Key
}

// GetCACert returns vpn server's cacert.
func (svr *Server) GetCACert() string {
	return svr.CACert
}

// GetCAKey returns vpn server's cakey.
func (svr *Server) GetCAKey() string {
	return svr.CAKey
}

// GetNet returns vpn server's net.
func (svr *Server) GetNet() string {
	return svr.Net
}

// GetMask returns vpn server's mask.
func (svr *Server) GetMask() string {
	return svr.Mask
}

// GetCRL returns vpn server's crl.
func (svr *Server) GetCRL() string {
	return svr.CRL
}

// GetDNS returns vpn server's dns.
func (svr *Server) GetDNS() string {
	if svr.DNS != "" {
		return svr.DNS
	}
	return DefaultVPNDNS
}

// GetCreatedAt returns server's created at.
func (svr *Server) GetCreatedAt() string {
	return svr.CreatedAt.Format(time.UnixDate)
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
func (svr *Server) Init(hostname string, port string, proto string, ipblock string, dns string) error {
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
	if svr := TheServer(); svr.IsInitialized() {
		if err := svr.Deinit(); err != nil {
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
	TheServer().EmitWithRestart()
	logrus.Infof("server initialized")
	return nil
}

// Update updates VPN server attributes.
func (svr *Server) Update(ipblock string, dns string) error {
	if !svr.IsInitialized() {
		return fmt.Errorf("server is not initialized")
	}

	var changed bool
	if ipblock != "" && govalidator.IsCIDR(ipblock) {
		var ipnet *net.IPNet
		_, ipnet, err := net.ParseCIDR(ipblock)
		if err != nil {
			return fmt.Errorf("can not parse CIDR %s: %v", ipblock, err)
		}
		svr.dbServerModel.Net = ipnet.IP.To4().String()
		svr.dbServerModel.Mask = net.IP(ipnet.Mask).To4().String()
		changed = true
	}

	if dns != "" && govalidator.IsIPv4(dns) {
		svr.dbServerModel.DNS = dns
		changed = true
	}
	if changed {
		db.Save(svr.dbServerModel)
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

		svr.EmitWithRestart()
		logrus.Infof("server updated")
	}
	return nil
}

// Deinit deletes the VPN server from the database and frees the allocated resources.
func (svr *Server) Deinit() error {
	if !svr.IsInitialized() {
		return fmt.Errorf("server not found")
	}

	db.Unscoped().Delete(&dbServerModel{})
	db.Unscoped().Delete(&dbRevokedModel{})
	svr.EmitWithRestart()
	return nil
}

// DumpsClientConfig generates .ovpn file for the given vpn user and returns it as a string.
func (svr *Server) DumpsClientConfig(username string) (string, error) {
	var result bytes.Buffer
	user, err := GetUser(username)
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
		Hostname: svr.GetHostname(),
		Port:     svr.GetPort(),
		CA:       svr.GetCACert(),
		Key:      user.getKey(),
		Cert:     user.GetCert(),
		NoGW:     user.IsNoGW(),
		Proto:    svr.GetProto(),
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
func (svr *Server) DumpClientConfig(username, path string) error {
	result, err := svr.DumpsClientConfig(username)
	if err != nil {
		return err
	}
	// Wite rendered content into openvpn server conf.
	return svr.emitToFile(path, result, 0)

}

// GetSystemCA returns the system CA from the database if available.
func (svr *Server) GetSystemCA() (*pki.CA, error) {
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
func (svr *Server) StartVPNProc() {
	if !svr.IsInitialized() {
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
	svr.Emit()
	vpnProc.Start()
	ensureNatEnabled()
}

// RestartVPNProc restarts the OpenVPN process.
func (svr *Server) RestartVPNProc() {
	if !svr.IsInitialized() {
		logrus.Error("can not launch OpenVPN because system is not initialized")
		return
	}
	if vpnProc == nil {
		panic(fmt.Sprintf("vpnProc is not initialized!"))
	}
	svr.Emit()
	vpnProc.Restart()
	ensureNatEnabled()
}

// StopVPNProc stops the OpenVPN process.
func (svr *Server) StopVPNProc() {
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
func (svr *Server) Emit() error {
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

	if !svr.IsInitialized() {
		return fmt.Errorf("you should create a server first. e.g. $ ovpm vpn create-server")
	}

	if err := svr.emitServerConf(); err != nil {
		return fmt.Errorf("can not emit server conf: %s", err)
	}

	if err := svr.emitServerCert(); err != nil {
		return fmt.Errorf("can not emit server cert: %s", err)
	}

	if err := svr.emitServerKey(); err != nil {
		return fmt.Errorf("can not emit server key: %s", err)
	}

	if err := svr.emitCACert(); err != nil {
		return fmt.Errorf("can not emit ca cert : %s", err)
	}

	if err := svr.emitCAKey(); err != nil {
		return fmt.Errorf("can not emit ca key: %s", err)
	}

	if err := svr.emitDHParams(); err != nil {
		return fmt.Errorf("can not emit dhparams: %s", err)
	}

	if err := svr.emitCCD(); err != nil {
		return fmt.Errorf("can not emit ccd: %s", err)
	}

	if err := svr.emitIptables(); err != nil {
		return fmt.Errorf("can not emit iptables: %s", err)
	}

	if err := svr.emitCRL(); err != nil {
		return fmt.Errorf("can not emit crl: %s", err)
	}

	logrus.Info("configurations emitted to the filesystem")
	return nil
}

// EmitWithRestart restarts vpnProc after calling EmitWithRestart().
func (svr *Server) EmitWithRestart() error {
	if err := svr.Emit(); err != nil {
		return err
	}
	if svr.IsInitialized() {
		for {
			if vpnProc.Status() == supervisor.RUNNING || vpnProc.Status() == supervisor.STOPPED {
				logrus.Info("OpenVPN process is restarting")
				svr.RestartVPNProc()
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	return nil

}

// emitToFile is a proxy that calls svr.emitToFileFunc.
func (svr *Server) emitToFile(path, content string, mode uint) error {
	return svr.emitToFileFunc(path, content, mode)
}

// emitToFile is an implementation for svr.emitToFileFunc.
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

func (svr *Server) emitServerConf() error {
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
		Net:          svr.Net,
		Mask:         svr.Mask,
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
	return svr.emitToFile(_DefaultVPNConfPath, result.String(), 0)
}

// Refresh synchronizes the server instance from db.
func (svr *Server) Refresh() error {
	//db = CreateDB("sqlite3", "")
	var dbServer dbServerModel

	q := db.First(&dbServer)
	if err := q.Error; err != nil {
		return fmt.Errorf("can't get server from db: %v", err)
	}
	if q.RecordNotFound() {
		return fmt.Errorf("server is not initialized")
	}
	svr.dbServerModel = dbServer
	return nil
}

// GetConnectedUsers will return a list of users who are currently connected
// to the VPN service.
func (svr *Server) GetConnectedUsers() ([]User, error) {
	var users []User

	// Open the status log file.
	f, err := svr.openFunc(_DefaultStatusLogPath)
	if err != nil {
		panic(err)
	}

	cl, _ := svr.parseStatusLogFunc(f) // client list from OpenVPN status log
	for _, c := range cl {
		var u dbUserModel
		q := db.Where(dbUserModel{Username: c.CommonName}).First(&u)
		if q.RecordNotFound() {
			logrus.WithFields(
				logrus.Fields{"CommonName": c.CommonName},
			).Error("user not found: common name reported by the OpenVPN is not found in the database")
			continue
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
func (svr *Server) IsInitialized() bool {
	var serverModel dbServerModel
	q := db.First(&serverModel)
	if err := q.Error; err != nil {
		logrus.Errorf("can't retrieve server from db: %v", err)
	}
	if q.RecordNotFound() {
		return false
	}
	return true
}

func (svr *Server) emitServerKey() error {
	// Write rendered content into key file.
	return svr.emitToFile(_DefaultKeyPath, svr.Key, 0600)
}

func (svr *Server) emitServerCert() error {
	// Write rendered content into the cert file.
	return svr.emitToFile(_DefaultCertPath, svr.Cert, 0)
}

func (svr *Server) emitCRL() error {
	var revokedDBItems []*dbRevokedModel
	db.Find(&revokedDBItems)
	var revokedCertSerials []*big.Int
	for _, item := range revokedDBItems {
		bi := big.NewInt(0)
		bi.SetString(item.SerialNumber, 16)
		revokedCertSerials = append(revokedCertSerials, bi)
	}
	systemCA, err := svr.GetSystemCA()
	if err != nil {
		return fmt.Errorf("can not emit CRL: %v", err)
	}
	crl, err := pki.NewCRL(systemCA, revokedCertSerials...)
	if err != nil {
		return fmt.Errorf("can not emit crl: %v", err)
	}

	return svr.emitToFile(_DefaultCRLPath, crl, 0)
}

func (svr *Server) emitCACert() error {
	// Write rendered content into the ca cert file.
	return svr.emitToFile(_DefaultCACertPath, svr.CACert, 0)
}

func (svr *Server) emitCAKey() error {
	// Write rendered content into the ca key file.
	return svr.emitToFile(_DefaultCAKeyPath, svr.CAKey, 0600)
}

func (svr *Server) emitCCD() error {
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
		}{IP: user.getIP().String(), NetMask: svr.Mask, Routes: associatedRoutes, Servernets: serverNets, RedirectGW: !user.NoGW}

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
		if err = svr.emitToFile(filepath.Join(_DefaultVPNCCDPath, user.Username), result.String(), 0); err != nil {
			return err
		}
	}
	return nil
}

func (svr *Server) emitDHParams() error {
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

	return svr.emitToFile(_DefaultDHParamsPath, result.String(), 0)
}

func (svr *Server) emitIptables() error {
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
