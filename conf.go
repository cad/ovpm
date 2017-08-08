package ovpm

const (
	// Version defines the version of ovpm.
	Version = "0.0.0"

	etcBasePath         = "/etc/ovpm/"
	varBasePath         = "/var/db/ovpm/"
	DefaultConfigPath   = etcBasePath + "ovpm.ini"
	DefaultDBPath       = varBasePath + "db.sqlite3"
	DefaultVPNConfPath  = varBasePath + "server.conf"
	DefaultVPNPort      = "1197"
	DefaultVPNCCDPath   = varBasePath + "ccd/"
	DefaultCertPath     = varBasePath + "server.crt"
	DefaultKeyPath      = varBasePath + "server.key"
	DefaultCACertPath   = varBasePath + "ca.crt"
	DefaultCAKeyPath    = varBasePath + "ca.key"
	DefaultDHParamsPath = varBasePath + "dh4096.pem"
	DefaultCRLPath      = varBasePath + "crl.pem"

	CrtExpireYears = 10
	CrtKeyLength   = 2024

	DefaultServerNetwork = "10.9.0.0"
	DefaultServerNetMask = "255.255.255.0"
)
