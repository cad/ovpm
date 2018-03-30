package ovpm

import (
	"fmt"
	"net"
	"reflect"
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

	// Test username validation.
	var networknametests = []struct {
		networkname string
		ok          bool
	}{
		{"asdf1240asfd", true},
		{"asdf.asfd", true},
		{"asdf12.12asfd", true},
		{"asd1f-as4fd", false},
		{"as0df a01sfd", false},
		{"as0df$a01sfd", false},
		{"as0df#a01sfd", false},
		{"a6sdf_as1fd", true},
	}

	for i, tt := range networknametests {
		_, err := CreateNewNetwork(tt.networkname, fmt.Sprintf("192.168.%d.0/24", i), SERVERNET, "")
		if ok := (err == nil); ok != tt.ok {
			t.Fatalf("expcted condition failed '%s': %v", tt.networkname, err)
		}
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

func TestNetworkTypeFromString(t *testing.T) {
	type args struct {
		typ string
	}
	tests := []struct {
		name string
		args args
		want NetworkType
	}{
		{"servernet", args{"SERVERNET"}, SERVERNET},
		{"route", args{"ROUTE"}, ROUTE},
		{"unknown", args{"aasdfsafdASDF"}, UNDEFINEDNET},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NetworkTypeFromString(tt.args.typ); got != tt.want {
				t.Errorf("NetworkTypeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllNetworkTypes(t *testing.T) {
	tests := []struct {
		name string
		want []NetworkType
	}{
		{"default", []NetworkType{UNDEFINEDNET, SERVERNET, ROUTE}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAllNetworkTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllNetworkTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetworkType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"servernet", args{"SERVERNET"}, true},
		{"route", args{"ROUTE"}, true},
		{"invalid", args{"ADSF"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNetworkType(tt.args.s); got != tt.want {
				t.Errorf("IsNetworkType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIncrementIP(t *testing.T) {
	type args struct {
		ip   string
		mask string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"default", args{"192.168.1.10", "255.255.255.0"}, "192.168.1.11", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementIP(tt.args.ip, tt.args.mask)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncrementIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_routableIP(t *testing.T) {
	type args struct {
		network string
		ip      net.IP
	}
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{"default", args{"ip4", net.ParseIP("0.0.0.0").To4()}, nil},
		{"local", args{"ip4", net.ParseIP("192.168.1.1").To4()}, net.ParseIP("192.168.1.1").To4()},
		{"v6linklocal", args{"ip6", net.ParseIP("FE80::").To16()}, net.ParseIP("FE80::").To16()},
		{"", args{"ip6", net.ParseIP("FE80::").To16()}, net.ParseIP("FE80::").To16()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routableIP(tt.args.network, tt.args.ip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("routableIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
