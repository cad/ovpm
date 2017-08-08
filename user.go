package ovpm

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
)

// User represents the interface that is being used within the public api.
type User interface {
	GetUsername() string
	GetServerSerialNumber() string
	GetCert() string
}

// DBUser is database model for VPN users.
type DBUser struct {
	gorm.Model
	ServerID uint
	Server   DBServer

	Username           string `gorm:"unique_index"`
	Cert               string
	ServerSerialNumber string
	Password           string
	Key                string
}

// DBRevoked is a database model for revoked VPN users.
type DBRevoked struct {
	gorm.Model
	SerialNumber string
}

func (u *DBUser) setPassword(newPassword string) error {
	// TODO(cad): Use a proper password hashing algorithm here.
	u.Password = newPassword
	return nil
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
// It also generates the necessary client keys and signs certificates with the current
// server's CA.
func CreateNewUser(username, password string) (*DBUser, error) {
	if !CheckBootstrapped() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(username) {
		return nil, fmt.Errorf("validation error: %s can not be null", username)
	}
	if !govalidator.IsAlphanumeric(username) {
		return nil, fmt.Errorf("validation error: `%s` can only contain letters and numbers", username)
	}
	ca, err := getCA()
	if err != nil {
		return nil, err
	}

	clientCert, err := CreateClientCert(username, ca)
	if err != nil {
		return nil, fmt.Errorf("can not create client cert %s: %v", username, err)
	}
	server, err := GetServerInstance()
	if err != nil {
		return nil, fmt.Errorf("can not get server: %v", err)
	}
	user := DBUser{

		Username:           username,
		Password:           password,
		Cert:               clientCert.Cert,
		Key:                clientCert.Key,
		ServerSerialNumber: server.SerialNumber,
	}

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
	crt, err := getCertFromPEM(u.Cert)
	if err != nil {
		return fmt.Errorf("can not get user's certificate: %v", err)
	}
	db.Create(&DBRevoked{
		SerialNumber: crt.SerialNumber.Text(16),
	})
	db.Unscoped().Delete(&u)

	err = Emit()
	if err != nil {
		return err
	}
	u = nil // delete the existing user struct
	return nil
}

// ResetPassword resets the users password into the provided password.
func (u *DBUser) ResetPassword(newPassword string) error {
	err := u.setPassword(newPassword)
	if err != nil {
		// user password can not be updated
		return fmt.Errorf("user password can not be updated %s: %v", u.Username, err)
	}
	db.Save(u)
	return nil
}

// Sign creates a key and a ceritificate signed by the current server's CA.
//
// This is often used to sign users when the current CA is changed while there are
// still  existing users in the database.
func (u *DBUser) Sign() error {
	if !CheckBootstrapped() {
		return fmt.Errorf("you first need to create server")
	}
	ca, err := getCA()
	if err != nil {
		return err
	}

	clientCert, err := CreateClientCert(u.Username, ca)
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
