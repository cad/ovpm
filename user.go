package ovpm

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
)

// GetUser finds and returns the user with the given username from database.
func GetUser(username string) (*User, error) {
	user := User{}
	db.Where(&User{Username: username}).First(&user)
	if db.NewRecord(&user) {
		// user is not found
		return nil, fmt.Errorf("user not found: %s", username)
	}
	return &user, nil
}

// GetAllUsers returns all recorded users in the database.
func GetAllUsers() ([]*User, error) {
	var users []*User
	db.Find(&users)

	return users, nil

}

// CreateUser creates a new user with the given username and password in the database.
// It also generates the necessary client keys and signs certificates with the current
// server's CA.
func CreateUser(username, password string) (*User, error) {
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

	user := User{
		Username: username,
		Password: password,
		Cert:     clientCert.Cert,
		Key:      clientCert.Key,
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

// DeleteUser deletes a user by the given username from the database.
func DeleteUser(username string) error {
	user := User{}
	db.Unscoped().Where(&User{Username: username}).First(&user)
	if db.NewRecord(&user) {
		// user is not found
		return fmt.Errorf("user not found: %s", username)
	}
	db.Unscoped().Delete(&user)

	err := Emit()
	if err != nil {
		return err
	}
	return nil
}

// ResetUserPassword resets the users password into the provided password.
func ResetUserPassword(username, newPassword string) error {
	user := User{}
	db.Where(&User{Username: username}).First(&user)
	if db.NewRecord(&user) {
		// user is not found
		return fmt.Errorf("user not found: %s", username)
	}

	err := user.setPassword(newPassword)
	if err != nil {
		// user password can not be updated
		return fmt.Errorf("user password can not be updated %s: %v", username, err)
	}
	return nil
}

// SignUser create a key and a ceritificate signed by the current server's CA.
//
// This is often used to sign users when the current CA is changed while there are
// still  existing users in the database.
func SignUser(username string) error {
	if !CheckBootstrapped() {
		return fmt.Errorf("you first need to create server")
	}
	user, err := GetUser(username)
	if err != nil {
		return fmt.Errorf("user not found %s: %v", username, err)
	}
	ca, err := getCA()
	if err != nil {
		return err
	}

	clientCert, err := CreateClientCert(username, ca)
	if err != nil {
		return fmt.Errorf("can not create client cert %s: %v", username, err)
	}

	server, err := GetServerInstance()
	if err != nil {
		return err
	}

	user.Cert = clientCert.Cert
	user.Key = clientCert.Key
	user.ServerSerialNumber = server.SerialNumber

	db.Save(&user)
	return nil
}
