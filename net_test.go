package ovpm

import "testing"

func TestVPNCreateNewNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"

	n, err := CreateNewNetwork(netName, cidrStr)
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	if n.Name != netName {
		t.Fatalf("network Name is expected to be '%s' but it's '%s' instead", netName, n.Name)
	}

	if n.CIDR != cidrStr {
		t.Fatalf("network CIDR is expected to be '%s' but it's '%s' instead", cidrStr, n.CIDR)
	}

	var network DBNetwork
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

}

func TestVPNDeleteNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"

	n, err := CreateNewNetwork(netName, cidrStr)
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	var network DBNetwork
	db.First(&network)

	if db.NewRecord(&network) {
		t.Fatalf("network is not created in the database.")
	}

	err = n.Delete()
	if err != nil {
		t.Fatalf("can't delete network: %v", err)
	}

	// Empty the existing network object.
	network = DBNetwork{}
	db.First(&network)
	if !db.NewRecord(&network) {
		t.Fatalf("network is not deleted from the database. %+v", network)
	}
}

func TestVPNGetNetwork(t *testing.T) {
	// Initialize:
	setupTestCase()
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	// Test:
	netName := "testnet"
	cidrStr := "192.168.1.0/24"

	_, err := CreateNewNetwork(netName, cidrStr)
	if err != nil {
		t.Fatalf("unexpected error when creating a new network: %v", err)
	}

	var network DBNetwork
	db.First(&network)

	if db.NewRecord(&network) {
		t.Fatalf("network is not created in the database.")
	}

	n, err := GetNetwork(netName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.NewRecord(&n) {
		t.Fatalf("network is not correctly returned from db.")
	}
}

func TestVPNGetAllNetworks(t *testing.T) {
	// Initialize:
	setupTestCase()
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	// Test:
	var getallnettests = []struct {
		name    string
		cidr    string
		passing bool
	}{
		{"testnet1", "192.168.1.0/24", true},
		{"testnet2", "10.10.0.0/16", true},
		{"testnet3", "asdkfjadflsa", false},
	}
	for _, tt := range getallnettests {
		_, err := CreateNewNetwork(tt.name, tt.cidr)
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
		}
	}
}

func init() {
	// Init
	Testing = true
}
