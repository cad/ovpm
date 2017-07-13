package ovpm

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

// User is database model for VPN users.
type User struct {
	gorm.Model
	ServerID           uint
	Server             Server
	ServerSerialNumber string

	Username string `gorm:"unique_index"`
	Password string
	Cert     string
	Key      string
}

func (u *User) setPassword(newPassword string) error {
	// TODO(cad): Use a proper password hashing algorithm here.
	u.Password = newPassword
	return nil
}

// Network is database model for external networks on the VPN server.
type Network struct {
	gorm.Model
	ServerID uint
	Server   Server

	Name        string
	NetworkCIDR string
}

// Server is database model for storing VPN server related stuff.
type Server struct {
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
}

// CheckSerial takes a serial number and checks it against the current server's serial number.
func (s *Server) CheckSerial(serialNo string) bool {
	return serialNo == s.SerialNumber
}

// CloseDB closes the database.
func CloseDB() {
	db.Close()
}

func init() {
	var err error
	db, err = gorm.Open("sqlite3", DefaultDBPath)
	if err != nil {
		logrus.Fatalf("couldn't open sqlite database %s: %v", DefaultDBPath, err)
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Network{})
	db.AutoMigrate(&Server{})
}
