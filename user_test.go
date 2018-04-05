package ovpm_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/cad/ovpm/pki"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
)

func TestCreateNewUser(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := ovpm.TheServer()
	svr.Init("localhost", "", ovpm.UDPProto, "", "")

	// Preare:
	username := "test.User"
	password := "testPasswd1234"
	noGW := false

	// Test:
	user, err := ovpm.CreateNewUser(username, password, noGW, 0, true)
	if err != nil {
		t.Fatalf("user can not be created: %v", err)
	}

	// Is user nil?
	if user == nil {
		t.Fatalf("user is expected to be 'NOT nil' but it is 'nil' %+v", user)
	}

	// Is user acutally exist in the system?
	user2, err := ovpm.GetUser(username)
	if err != nil {
		t.Fatalf("user can not be retrieved: %v", err)
	}

	// Are users the same?
	if !areUsersEqual(user, user2) {
		t.Fatalf("users are expected to be 'NOT TO DIFFER' but they are 'DIFFERENT'")
	}

	// Is user's server serial number correct?
	if !svr.CheckSerial(user.ServerSerialNumber) {
		t.Fatalf("user's ServerSerialNumber is expected to be 'CORRECT' but it is 'INCORRECT' instead %+v", user)
	}

	// Does User interface work properly?
	if user.GetUsername() != user.Username {
		t.Errorf("user.GetUsername() is expected to return '%s' but it returns '%s' %+v", user.Username, user.GetUsername(), user)
	}
	if user.GetServerSerialNumber() != user.ServerSerialNumber {
		t.Errorf("user.GetServerSerialNumber() is expected to return '%s' but it returns '%s' %+v", user.ServerSerialNumber, user.GetServerSerialNumber(), user)
	}
	if user.GetCert() != user.Cert {
		t.Errorf("user.GetCert() is expected to return '%s' but it returns '%s' %+v", user.Cert, user.GetCert(), user)
	}

	user.Delete()

	// Is NoGW attr working properly?
	noGW = true
	user, err = ovpm.CreateNewUser(username, password, noGW, 0, true)
	if err != nil {
		t.Fatalf("user can not be created: %v", err)
	}
	if user.NoGW != noGW {
		t.Fatalf("user.NoGW is expected to be %t but it's %t instead", noGW, user.NoGW)
	}

	// Try to create a user with an invalid static ip.
	user = nil
	_, err = ovpm.CreateNewUser("staticuser", password, noGW, ovpm.IP2HostID(net.ParseIP("8.8.8.8").To4()), true)
	if err == nil {
		t.Fatalf("user creation expected to err but it didn't")
	}

	// Test username validation.
	var usernametests = []struct {
		username string
		ok       bool
	}{
		{"asdf1240asfd", true},
		{"asdf.asfd", true},
		{"asdf12.12asfd", true},
		{"asd1f-as4fd", false},
		{"as0df a01sfd", false},
		{"a6sdf_as1fd", true},
	}

	for _, tt := range usernametests {
		_, err := ovpm.CreateNewUser(tt.username, "1234", false, 0, true)
		if ok := (err == nil); ok != tt.ok {
			t.Fatalf("expcted condition failed '%s': %v", tt.username, err)
		}
	}
}

func TestUserUpdate(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	username := "testUser"
	password := "testPasswd1234"
	noGW := false

	// Test:
	user, err := ovpm.CreateNewUser(username, password, noGW, 0, true)
	if err != nil {
		t.Fatalf("user can not be created: %v", err)
	}

	var updatetests = []struct {
		password string
		noGW     bool
		hostid   uint32
		ok       bool
	}{
		{"testpw", false, 0, true},
		{"123", false, 0, true},
		{"123", false, 0, true},
		{"", true, 0, true},
		{"", true, ovpm.IP2HostID(net.ParseIP("10.10.10.10").To4()), false}, // Invalid static address.
		{"333", true, ovpm.IP2HostID(net.ParseIP("10.9.0.7").To4()), true},
		{"222", true, ovpm.IP2HostID(net.ParseIP("10.9.0.7").To4()), true},
	}

	for _, tt := range updatetests {
		err := user.Update(tt.password, tt.noGW, tt.hostid, true)
		if (err == nil) != tt.ok {
			t.Errorf("user is expected to be able to update but it gave us this error instead: %v", err)
		}
	}
}

func TestUserPasswordCorrect(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	initialPassword := "g00dp@ssW0rd9"
	user, _ := ovpm.CreateNewUser("testUser", initialPassword, false, 0, true)

	// Test:
	// Is user created with the correct password?
	if !user.CheckPassword(initialPassword) {
		t.Fatalf("user's password must be '%s', but CheckPassword fails +%v", initialPassword, user)
	}
}

func TestUserPasswordReset(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	initialPassword := "g00dp@ssW0rd9"
	user, _ := ovpm.CreateNewUser("testUser", initialPassword, false, 0, true)

	// Test:

	// Reset user's password.
	newPassword := "@n0th3rP@ssw0rd"
	user.ResetPassword(newPassword)

	// Is newPassword set correctly?
	if !user.CheckPassword(newPassword) {
		t.Fatalf("user's password must be '%s', but CheckPassword fails +%v", newPassword, user)
	}

	// Is initialPassword is invalid now?
	if user.CheckPassword(initialPassword) {
		t.Fatalf("user's password must be '%s', but CheckPassword returns true for the old password '%s' %+v", newPassword, initialPassword, user)
	}
}

func TestUserDelete(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	username := "testUser"
	user, _ := ovpm.CreateNewUser(username, "1234", false, 0, true)

	// Test:

	// Is user fetchable?
	_, err := ovpm.GetUser(username)
	if err != nil {
		t.Fatalf("user '%s' expected to be 'EXIST', but we failed to get it %+v: %v", username, user, err)
	}

	// Delete the user.
	err = user.Delete()

	// Is user deleted?
	if err != nil {
		t.Fatalf("user is expected to be deleted, but we got this error instead: %v", err)
	}

	// Is user now fetchable? (It shouldn't be.)
	u1, err := ovpm.GetUser(username)
	if err == nil {
		t.Fatalf("user '%s' expected to be 'NOT EXIST', but it does exist instead: %v", username, user)
	}

	// Is user nil?
	if u1 != nil {
		t.Fatalf("not found user should be 'nil' but it's '%v' instead", u1)
	}
}

func TestUserGet(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	username := "testUser"
	user, _ := ovpm.CreateNewUser(username, "1234", false, 0, true)

	// Test:
	// Is user fetchable?
	fetchedUser, err := ovpm.GetUser(username)
	if err != nil {
		t.Fatalf("user should be fetchable but instead we got this error: %v", err)
	}

	// Is fetched user same with the created one?
	if !areUsersEqual(user, fetchedUser) {
		t.Fatalf("fetched user should be same with the created user but it differs")
	}

}

func TestUserGetAll(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	ovpm.TheServer().Init("localhost", "", ovpm.UDPProto, "", "")
	count := 5

	// Prepare:
	var users []*ovpm.User
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("user%d", i)
		password := fmt.Sprintf("password%d", i)
		user, _ := ovpm.CreateNewUser(username, password, false, 0, true)
		users = append(users, user)
	}

	// Test:
	// Get all users.
	fetchedUsers, err := ovpm.GetAllUsers()

	// Is users are fetchable.
	if err != nil {
		t.Fatalf("users are expected to be fetchable but we got this error instead: %v", err)
	}

	// Is the fetched user count is correct?
	if len(fetchedUsers) != count {
		t.Fatalf("fetched user count is expected to be '%d', but it is '%d' instead", count, len(fetchedUsers))
	}

	// Are returned users same with the created ones?
	for i, user := range fetchedUsers {
		if !areUsersEqual(user, users[i]) {
			t.Fatalf("user %s[%d] is expected to be 'SAME' with the created one of it, but they are 'DIFFERENT'", user.GetUsername(), i)
		}
	}
}

func TestUserRenew(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := ovpm.TheServer()
	svr.Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:
	user, _ := ovpm.CreateNewUser("user", "1234", false, 0, true)

	// Test:
	// Re initialize the server.
	svr.Init("example.com", "3333", ovpm.UDPProto, "", "") // This causes implicit Renew() on every user in the system.

	// Fetch user back.
	fetchedUser, _ := ovpm.GetUser(user.GetUsername())

	// Aren't users certificates different?
	if user.Cert == fetchedUser.Cert {
		t.Fatalf("fetched user's certificate is expected to be 'DIFFERENT' from the created user (since server is re initialized), but it's 'SAME' instead")
	}
}

func TestUserIPAllocator(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := ovpm.TheServer()
	svr.Init("localhost", "", ovpm.UDPProto, "", "")

	// Prepare:

	// Test:
	var iptests = []struct {
		username   string
		gw         bool
		hostid     uint32
		expectedIP string
		pass       bool
	}{
		{"user1", false, 0, "10.9.0.2/24", true},
		{"user2", false, 0, "10.9.0.3/24", true},
		{"user3", true, 0, "10.9.0.4/24", true},
		{"user4", true, ovpm.IP2HostID(net.ParseIP("10.9.0.5").To4()), "10.9.0.5/24", true},
		{"user6", true, ovpm.IP2HostID(net.ParseIP("10.9.0.7").To4()), "10.9.0.7/24", true},
		{"user7", true, 0, "10.9.0.6/24", true},
		{"user6", true, ovpm.IP2HostID(net.ParseIP("10.9.0.1").To4()), "10.9.0.7/24", false},
	}
	for _, tt := range iptests {
		user, err := ovpm.CreateNewUser(tt.username, "pass", tt.gw, tt.hostid, true)
		if (err == nil) == !tt.pass {
			t.Fatalf("expected pass %t %s", tt.pass, err)
		}
		if user != nil {
			if user.GetIPNet() != tt.expectedIP {
				t.Fatalf("user %s ip %s(%d) is expected to be %s", user.GetUsername(), user.GetIPNet(), user.GetHostID(), tt.expectedIP)
			}
		}
	}
}
func TestUser_ExpiresAt(t *testing.T) {
	// Initialize:
	db := ovpm.CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := ovpm.TheServer()
	svr.Init("localhost", "", ovpm.UDPProto, "", "")

	// Test:
	u1, err := ovpm.CreateNewUser("test", "1234", true, 0, false)
	if err != nil {
		t.Fatalf("test preperation failed: %v", err)
	}

	cert, err := pki.ReadCertFromPEM(u1.Cert)
	if err != nil {
		t.Fatalf("test preperation failed: %v", err)
	}

	if !reflect.DeepEqual(u1.ExpiresAt(), cert.NotAfter) {
		t.Errorf("got (%s), want (%s)", u1.ExpiresAt(), cert.NotAfter)
	}
}

// areUsersEqual compares given users and returns true if they are the same.
func areUsersEqual(user1, user2 *ovpm.User) bool {
	if user1.GetCert() != user2.GetCert() {
		logrus.Info("Cert %v != %v", user1.GetCert(), user2.GetCert())
		return false
	}
	if user1.GetUsername() != user2.GetUsername() {
		logrus.Infof("Username %v != %v", user1.GetUsername(), user2.GetUsername())
		return false
	}

	if user1.GetServerSerialNumber() != user2.GetServerSerialNumber() {
		logrus.Infof("ServerSerialNumber %v != %v", user1.GetServerSerialNumber(), user2.GetServerSerialNumber())
		return false
	}
	logrus.Infof("users are the same!")
	return true
}

func init() {
	ovpm.Testing = true
}
