package ovpm

const (
	// Version defines the version of ovpm.
	Version = "0.1.10"

	// DefaultVPNPort is the default OpenVPN port to listen.
	DefaultVPNPort = "1197"

	etcBasePath = "/etc/ovpm/"
	varBasePath = "/var/db/ovpm/"

	_DefaultConfigPath   = etcBasePath + "ovpm.ini"
	_DefaultDBPath       = varBasePath + "db.sqlite3"
	_DefaultVPNConfPath  = varBasePath + "server.conf"
	_DefaultVPNCCDPath   = varBasePath + "ccd/"
	_DefaultCertPath     = varBasePath + "server.crt"
	_DefaultKeyPath      = varBasePath + "server.key"
	_DefaultCACertPath   = varBasePath + "ca.crt"
	_DefaultCAKeyPath    = varBasePath + "ca.key"
	_DefaultDHParamsPath = varBasePath + "dh4096.pem"
	_DefaultCRLPath      = varBasePath + "crl.pem"

	_DefaultServerNetwork = "10.9.0.0"
	_DefaultServerNetMask = "255.255.255.0"
)

// Testing is used to determine wether we are testing or running normally.
// Set it to true when testing.
var Testing = false
