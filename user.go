package ovpm

import (
	"fmt"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm/pki"
	"github.com/jinzhu/gorm"
	"gopkg.in/hlandau/passlib.v1"
)

// User represents the interface that is being used within the public api.
type User interface {
	GetUsername() string
	GetServerSerialNumber() string
	GetCert() string
	GetIPNet() string
	IsNoGW() bool
}

// DBUser is database model for VPN users.
type DBUser struct {
	gorm.Model
	ServerID uint
	Server   DBServer

	Username           string `gorm:"unique_index"`
	Cert               string
	ServerSerialNumber string
	Hash               string
	Key                string
	NoGW               bool
}

// DBRevoked is a database model for revoked VPN users.
type DBRevoked struct {
	gorm.Model
	SerialNumber string
}

func (u *DBUser) setPassword(password string) error {
	hashedPassword, err := passlib.Hash(password)
	if err != nil {
		return fmt.Errorf("can not set password: %v", err)
	}

	u.Hash = hashedPassword
	return nil
}

// CheckPassword returns wether the given password is correct for the user.
func (u *DBUser) CheckPassword(password string) bool {
	_, err := passlib.Verify(password, u.Hash)
	if err != nil {
		logrus.Error(err)
		return false
	}
	return true
}

// GetUser finds and returns the user with the given username from database.
func GetUser(username string) (*DBUser, error) {
	user := DBUser{}
	db.Where(&DBUser{Username: username}).First(&user)
	if db.NewRecord(&user) {
		// user is not found
		return nil, fmt.Errorf("user not found: %s", username)
	}
	return &user, nil
}

// GetAllUsers returns all recorded users in the database.
func GetAllUsers() ([]*DBUser, error) {
	var users []*DBUser
	db.Find(&users)

	return users, nil

}

// CreateNewUser creates a new user with the given username and password in the database.
// If nogw is true, then ovpm doesn't push vpn server as the default gw for the user.
//
// It also generates the necessary client keys and signs certificates with the current
// server's CA.
func CreateNewUser(username, password string, nogw bool) (*DBUser, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(username) {
		return nil, fmt.Errorf("validation error: %s can not be null", username)
	}
	if !govalidator.IsAlphanumeric(username) {
		return nil, fmt.Errorf("validation error: `%s` can only contain letters and numbers", username)
	}
	ca, err := GetSystemCA()
	if err != nil {
		return nil, err
	}

	clientCert, err := pki.NewClientCertHolder(ca, username)
	if err != nil {
		return nil, fmt.Errorf("can not create client cert %s: %v", username, err)
	}
	server, err := GetServerInstance()
	if err != nil {
		return nil, fmt.Errorf("can not get server: %v", err)
	}
	user := DBUser{
		Username:           username,
		Cert:               clientCert.Cert,
		Key:                clientCert.Key,
		ServerSerialNumber: server.SerialNumber,
		NoGW:               nogw,
	}
	user.setPassword(password)

	db.Create(&user)
	if db.NewRecord(&user) {
		// user is still not created
		return nil, fmt.Errorf("can not create user in database: %s", user.Username)
	}
	logrus.Infof("user created: %s", username)

	// Emit server config
	err = Emit()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete deletes a user by the given username from the database.
func (u *DBUser) Delete() error {
	if db.NewRecord(&u) {
		// user is not found
		return fmt.Errorf("user is not initialized: %s", u.Username)
	}
	crt, err := pki.ReadCertFromPEM(u.Cert)
	if err != nil {
		return fmt.Errorf("can not get user's certificate: %v", err)
	}
	db.Create(&DBRevoked{
		SerialNumber: crt.SerialNumber.Text(16),
	})
	db.Unscoped().Delete(&u)
	logrus.Infof("user deleted: %s", u.GetUsername())
	err = Emit()
	if err != nil {
		return err
	}
	u = nil // delete the existing user struct
	return nil
}

// ResetPassword resets the users password into the provided password.
func (u *DBUser) ResetPassword(password string) error {
	err := u.setPassword(password)
	if err != nil {
		// user password can not be updated
		return fmt.Errorf("user password can not be updated %s: %v", u.Username, err)
	}
	db.Save(u)
	err = Emit()
	if err != nil {
		return err
	}

	logrus.Infof("user password reset: %s", u.GetUsername())
	return nil
}

// Renew creates a key and a ceritificate signed by the current server's CA.
//
// This is often used to sign users when the current CA is changed while there are
// still  existing users in the database.
func (u *DBUser) Renew() error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}
	ca, err := GetSystemCA()
	if err != nil {
		return err
	}

	clientCert, err := pki.NewClientCertHolder(ca, u.Username)
	if err != nil {
		return fmt.Errorf("can not create client cert %s: %v", u.Username, err)
	}

	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	u.Cert = clientCert.Cert
	u.Key = clientCert.Key
	u.ServerSerialNumber = server.SerialNumber

	db.Save(&u)
	err = Emit()
	if err != nil {
		return err
	}

	logrus.Infof("user renewed cert: %s", u.GetUsername())
	return nil
}

// GetUsername returns user's username.
func (u *DBUser) GetUsername() string {
	return u.Username
}

// GetCert returns user's public certificate.
func (u *DBUser) GetCert() string {
	return u.Cert
}

// GetServerSerialNumber returns user's server serial number.
func (u *DBUser) GetServerSerialNumber() string {
	return u.ServerSerialNumber
}

// GetCreatedAt returns user's creation time.
func (u *DBUser) GetCreatedAt() string {
	return u.CreatedAt.Format(time.UnixDate)
}

// getIP returns user's vpn ip addr.
func (u *DBUser) getIP() net.IP {
	clientsNetMask := net.IPMask(net.ParseIP(_DefaultServerNetMask))
	clientsNetPrefix := net.ParseIP(_DefaultServerNetwork)
	clientNet := clientsNetPrefix.Mask(clientsNetMask).To4()
	clientNet[3] = byte(u.ID)
	return clientNet
}

// GetIPNet returns user's vpn ip network. (e.g. 192.168.0.1/24)
func (u *DBUser) GetIPNet() string {
	mask := net.IPMask(net.ParseIP(_DefaultServerNetMask))

	ipn := net.IPNet{
		IP:   u.getIP(),
		Mask: mask,
	}
	return ipn.String()
}

// IsNoGW returns wether user is set to get the vpn server as their default gateway.
func (u *DBUser) IsNoGW() bool {
	return u.NoGW
}
