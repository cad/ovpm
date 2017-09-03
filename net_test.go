package ovpm

import (
	"testing"
)

func TestVPNCreateNewNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET

	n, err := CreateNewNetwork(netName, cidrStr, netType, "")
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	if n.Name != netName {
		t.Fatalf("network Name is expected to be '%s' but it's '%s' instead", netName, n.Name)
	}

	if n.CIDR != cidrStr {
		t.Fatalf("network CIDR is expected to be '%s' but it's '%s' instead", cidrStr, n.CIDR)
	}

	var network dbNetworkModel
	db.First(&network)

	if db.NewRecord(&network) {
		t.Fatalf("network is not created in the database.")
	}

	if network.Name != netName {
		t.Fatalf("network Name is expected to be '%s' but it's '%s' instead", netName, network.Name)
	}

	if network.CIDR != cidrStr {
		t.Fatalf("network CIDR is expected to be '%s' but it's '%s' instead", cidrStr, network.CIDR)
	}

	if network.Type != netType {
		t.Fatalf("network CIDR is expected to be '%s' but it's '%s' instead", netType, network.Type)
	}

}

func TestVPNDeleteNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET

	n, err := CreateNewNetwork(netName, cidrStr, netType, "")
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	var network dbNetworkModel
	db.First(&network)

	if db.NewRecord(&network) {
		t.Fatalf("network is not created in the database.")
	}

	err = n.Delete()
	if err != nil {
		t.Fatalf("can't delete network: %v", err)
	}

	// Empty the existing network object.
	network = dbNetworkModel{}
	db.First(&network)
	if !db.NewRecord(&network) {
		t.Fatalf("network is not deleted from the database. %+v", network)
	}
}

func TestVPNGetNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET

	_, err := CreateNewNetwork(netName, cidrStr, netType, "")
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	var network dbNetworkModel
	db.First(&network)

	if db.NewRecord(&network) {
		t.Fatalf("network is not created in the database.")
	}

	n, err := GetNetwork(netName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.NewRecord(&n.dbNetworkModel) {
		t.Fatalf("network is not correctly returned from db.")
	}
}

func TestVPNGetAllNetworks(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	var getallnettests = []struct {
		name    string
		cidr    string
		netType NetworkType
		passing bool
	}{
		{"testnet1", "192.168.1.0/24", SERVERNET, true},
		{"testnet2", "10.10.0.0/16", SERVERNET, true},
		{"testnet3", "asdkfjadflsa", SERVERNET, false},
	}
	for _, tt := range getallnettests {
		_, err := CreateNewNetwork(tt.name, tt.cidr, tt.netType, "")
		if (err == nil) != tt.passing {
			t.Fatalf("unexpected error when creating a new network: %v", err)
		}
	}

	for _, tt := range getallnettests {
		n, err := GetNetwork(tt.name)
		if (err == nil) != tt.passing {
			t.Fatalf("network's presence is expected to be '%t' but it's '%t' instead", tt.passing, !tt.passing)
		}

		if tt.passing {
			if n.Name != tt.name {
				t.Fatalf("network Name is expected to be '%s' but it's '%s'", tt.name, n.Name)
			}
			if n.CIDR != tt.cidr {
				t.Fatalf("network CIDR is expected to be '%s' but it's '%s'", tt.cidr, n.CIDR)
			}
			if n.Type != tt.netType {
				t.Fatalf("network CIDR is expected to be '%s' but it's '%s' instead", tt.netType, n.Type)
			}
		}
	}
}

func TestNetAssociate(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET
	userName := "testUser2"
	user, err := CreateNewUser(userName, "123", false, 0, true)
	if err != nil {
		t.Fatal(err)
	}

	n, err := CreateNewNetwork(netName, cidrStr, netType, "")
	if err != nil {
		t.Fatal(err)
	}

	err = n.Associate(user.dbUserModel.Username)
	if err != nil {
		t.Fatal(err)
	}
	n = nil

	n, err = GetNetwork(netName)
	if err != nil {
		t.Fatal(err)
	}

	// Does number of associated users in the network object matches the number that we have created?
	if count := len(n.dbNetworkModel.Users); count != 1 {
		t.Fatalf("network.Users count is expexted to be %d, but it's %d", 1, count)
	}
	err = n.Associate(user.dbUserModel.Username)
	if err == nil {
		t.Fatalf("expected to get error but got no error instead")
	}

}

func TestNetDissociate(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	err := Init("localhost", "", UDPProto, "", "")
	if err != nil {
		t.Fatal(err)
	}

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET
	userName := "testUser2"
	user, err := CreateNewUser(userName, "123", false, 0, true)
	if err != nil {
		t.Fatal(err)
	}

	n, err := CreateNewNetwork(netName, cidrStr, netType, "")
	if err != nil {
		t.Fatal(err)
	}
	n.Associate(user.Username)

	n = nil
	n, _ = GetNetwork(netName)

	// Does number of associated users in the network object matches the number that we have created?
	if count := len(n.Users); count != 1 {
		t.Fatalf("network.Users count is expexted to be %d, but it's %d", 1, count)
	}

	err = n.Dissociate(user.Username)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n = nil
	n, _ = GetNetwork(netName)

	// Does number of associated users in the network object matches the number that we have created?
	if count := len(n.Users); count != 0 {
		t.Fatalf("network.Users count is expexted to be %d, but it's %d", 0, count)
	}
	err = n.Dissociate(user.Username)
	if err == nil {
		t.Fatalf("expected error but got no error instead")
	}
}

func TestNetGetAssociatedUsers(t *testing.T) {
	// Initialize:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	Init("localhost", "", UDPProto, "", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"
	netType := SERVERNET
	userName := "testUser2"
	user, _ := CreateNewUser(userName, "123", false, 0, true)

	n, _ := CreateNewNetwork(netName, cidrStr, netType, "")
	n.Associate(user.Username)
	n = nil
	n, _ = GetNetwork(netName)

	// Test:
	if n.GetAssociatedUsers()[0].Username != user.Username {
		t.Fatalf("returned associated user is expected to be the same user with the one we have created, but its not")
	}
}

func init() {
	// Init
	Testing = true
}
