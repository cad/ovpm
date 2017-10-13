package ovpm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/coreos/go-iptables/iptables"
	"github.com/jinzhu/gorm"
)

// NetworkType distinguishes different types of networks that is defined in the networks table.
type NetworkType uint

// NetworkTypes
const (
	UNDEFINEDNET NetworkType = iota
	SERVERNET
	ROUTE
)

var networkTypes = [...]struct {
	Type        NetworkType
	String      string
	Description string
}{
	{UNDEFINEDNET, "UNDEFINEDNET", "unknown network type"},
	{SERVERNET, "SERVERNET", "network behind vpn server"},
	{ROUTE, "ROUTE", "network to be pushed as route"},
}

// NetworkTypeFromString returns string representation of the network type.
func NetworkTypeFromString(typ string) NetworkType {
	for _, v := range networkTypes {
		if v.String == typ {
			return v.Type
		}
	}
	return UNDEFINEDNET
}

// GetAllNetworkTypes returns all network types defined in the system.
func GetAllNetworkTypes() []NetworkType {
	var networkTypeList []NetworkType
	for _, v := range networkTypes {
		networkTypeList = append(networkTypeList, v.Type)
	}
	return networkTypeList
}

func (nt NetworkType) String() string {
	for _, v := range networkTypes {
		if v.Type == nt {
			return v.String
		}
	}
	return "UNDEFINEDNET"
}

// Description gives description about the network type.
func (nt NetworkType) Description() string {
	for _, v := range networkTypes {
		if v.Type == nt {
			return v.Description
		}
	}
	return "UNDEFINEDNET"
}

// dbNetworkModel is database model for external networks on the VPN server.
type dbNetworkModel struct {
	gorm.Model
	ServerID uint
	Server   dbServerModel

	Name  string `gorm:"unique_index"`
	CIDR  string
	Type  NetworkType
	Via   string
	Users []*dbUserModel `gorm:"many2many:network_users;"`
}

// Network represents a VPN related network.
type Network struct {
	dbNetworkModel
}

// GetNetwork returns a network specified by its name.
func GetNetwork(name string) (*Network, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(name) {
		return nil, fmt.Errorf("validation error: %s can not be null", name)
	}

	var network dbNetworkModel
	db.Preload("Users").Where(&dbNetworkModel{Name: name}).First(&network)

	if db.NewRecord(&network) {
		return nil, fmt.Errorf("network not found %s", name)
	}

	return &Network{dbNetworkModel: network}, nil
}

// GetAllNetworks returns all networks defined in the system.
func GetAllNetworks() []*Network {
	var networks []*Network
	var dbNetworks []*dbNetworkModel
	db.Preload("Users").Find(&dbNetworks)
	for _, n := range dbNetworks {
		networks = append(networks, &Network{dbNetworkModel: *n})
	}
	return networks
}

// CreateNewNetwork creates a new network definition in the system.
func CreateNewNetwork(name, cidr string, nettype NetworkType, via string) (*Network, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(name) {
		return nil, fmt.Errorf("validation error: %s can not be null", name)
	}
	if !govalidator.Matches(name, "^([\\w\\.]+)$") { // allow alphanumeric, underscore and dot
		return nil, fmt.Errorf("validation error: `%s` can only contain letters, numbers, underscores and dots", name)
	}
	if !govalidator.IsCIDR(cidr) {
		return nil, fmt.Errorf("validation error: `%s` must be a network in the CIDR form", cidr)
	}

	if via != "" && !govalidator.IsIPv4(via) {
		return nil, fmt.Errorf("validation error: `%s` must be a network in the IPv4 form", via)
	}

	if nettype == UNDEFINEDNET {
		return nil, fmt.Errorf("validation error: `%s` must be a valid network type", nettype)
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("can not parse CIDR %s: %v", cidr, err)
	}

	// Overwrite via with the parsed IPv4 string.
	if nettype == ROUTE && via != "" {
		viaIP := net.ParseIP(via).To4()
		if err != nil {
			return nil, fmt.Errorf("can not parse IPv4 %s: %v", via, err)
		}
		via = viaIP.String()

	} else {
		via = ""
	}

	network := dbNetworkModel{
		Name:  name,
		CIDR:  ipnet.String(),
		Type:  nettype,
		Users: []*dbUserModel{},
		Via:   via,
	}
	db.Save(&network)

	if db.NewRecord(&network) {
		return nil, fmt.Errorf("can not create network in the db")
	}
	EmitWithRestart()
	logrus.Infof("network defined: %s (%s)", network.Name, network.CIDR)
	return &Network{dbNetworkModel: network}, nil

}

// Delete deletes a network definition in the system.
func (n *Network) Delete() error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}

	db.Unscoped().Delete(n.dbNetworkModel)
	EmitWithRestart()
	logrus.Infof("network deleted: %s", n.Name)
	return nil
}

// Associate allows the given user access to this network.
func (n *Network) Associate(username string) error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}
	user, err := GetUser(username)
	if err != nil {
		return fmt.Errorf("user can not be fetched: %v", err)
	}

	var users []dbUserModel
	userAssoc := db.Model(&n.dbNetworkModel).Association("Users")
	userAssoc.Find(&users)
	var found bool
	for _, u := range users {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if found {
		return fmt.Errorf("user %s is already associated with the network %s", user.Username, n.Name)
	}

	userAssoc.Append(user.dbUserModel)
	if userAssoc.Error != nil {
		return fmt.Errorf("association failed: %v", userAssoc.Error)
	}
	EmitWithRestart()
	logrus.Infof("user '%s' is associated with the network '%s'", user.GetUsername(), n.Name)
	return nil
}

// Dissociate breaks up the given users association to the said network.
func (n *Network) Dissociate(username string) error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}

	user, err := GetUser(username)
	if err != nil {
		return fmt.Errorf("user can not be fetched: %v", err)
	}

	var users []dbUserModel
	userAssoc := db.Model(&n.dbNetworkModel).Association("Users")
	userAssoc.Find(&users)
	var found bool
	for _, u := range users {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("user %s is already not associated with the network %s", user.Username, n.Name)
	}

	userAssoc.Delete(user.dbUserModel)
	if userAssoc.Error != nil {
		return fmt.Errorf("disassociation failed: %v", userAssoc.Error)
	}
	EmitWithRestart()
	logrus.Infof("user '%s' is dissociated with the network '%s'", user.GetUsername(), n.Name)
	return nil
}

// GetName returns network's name.
func (n *Network) GetName() string {
	return n.Name
}

// GetCIDR returns network's CIDR.
func (n *Network) GetCIDR() string {
	return n.CIDR
}

// GetCreatedAt returns network's name.
func (n *Network) GetCreatedAt() string {
	return n.CreatedAt.Format(time.UnixDate)
}

// GetType returns network's network type.
func (n *Network) GetType() NetworkType {
	return NetworkType(n.Type)
}

// GetAssociatedUsers returns network's associated users.
func (n *Network) GetAssociatedUsers() []*User {
	var users []*User
	for _, u := range n.Users {
		users = append(users, &User{dbUserModel: *u})
	}
	return users
}

// GetAssociatedUsernames returns network's associated user names.
func (n *Network) GetAssociatedUsernames() []string {
	var usernames []string

	for _, user := range n.GetAssociatedUsers() {
		usernames = append(usernames, user.Username)
	}
	return usernames
}

// GetVia returns network' via.
func (n *Network) GetVia() string {
	return n.Via
}

// interfaceOfIP returns a network interface that has the given IP.
func interfaceOfIP(ipnet *net.IPNet) *net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logrus.Error(err)
			return nil
		}
		for _, addr := range addrs {
			switch addr := addr.(type) {
			case *net.IPAddr:
				if ip := addr.IP; ip != nil {
					if ipnet.Contains(ip) {
						return &iface
					}
				}
			case *net.IPNet:
				if ip := addr.IP; ip != nil {
					if ipnet.Contains(ip) {
						return &iface
					}
				}
			}
		}
	}
	return nil
}

func routedInterface(network string, flags net.Flags) *net.Interface {
	switch network {
	case "ip", "ip4", "ip6":
	default:
		return nil
	}
	ift, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ift {
		if ifi.Flags&flags != flags {
			continue
		}
		if _, ok := hasRoutableIP(network, &ifi); !ok {
			continue
		}
		return &ifi
	}
	return nil
}

func getOutboundInterface() *net.Interface {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ipnet := net.IPNet{
		IP:   localAddr.IP.To4(),
		Mask: localAddr.IP.To4().DefaultMask(),
	}
	return interfaceOfIP(&ipnet)
}

func hasRoutableIP(network string, ifi *net.Interface) (net.IP, bool) {
	ifat, err := ifi.Addrs()
	if err != nil {
		return nil, false
	}
	for _, ifa := range ifat {
		switch ifa := ifa.(type) {
		case *net.IPAddr:
			if ip := routableIP(network, ifa.IP); ip != nil {
				return ip, true
			}
		case *net.IPNet:
			if ip := routableIP(network, ifa.IP); ip != nil {
				return ip, true
			}
		}
	}
	return nil, false
}

func vpnInterface() *net.Interface {
	server, err := GetServerInstance()
	if err != nil {
		logrus.Errorf("can't get server instance: %v", err)
		return nil
	}

	mask := net.IPMask(net.ParseIP(server.Mask))
	prefix := net.ParseIP(server.Net)
	netw := prefix.Mask(mask).To4()
	netw[3] = byte(1) // Server is always gets xxx.xxx.xxx.1
	ipnet := net.IPNet{IP: netw, Mask: mask}

	return interfaceOfIP(&ipnet)
}

func routableIP(network string, ip net.IP) net.IP {
	if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsGlobalUnicast() {
		return nil
	}
	switch network {
	case "ip4":
		if ip := ip.To4(); ip != nil {
			return ip
		}
	case "ip6":
		if ip.IsLoopback() { // addressing scope of the loopback address depends on each implementation
			return nil
		}
		if ip := ip.To16(); ip != nil && ip.To4() == nil {
			return ip
		}
	default:
		if ip := ip.To4(); ip != nil {
			return ip
		}
		if ip := ip.To16(); ip != nil {
			return ip
		}
	}
	return nil
}

// ensureNatEnabled launches a goroutine that constantly tries to enable nat.
func ensureNatEnabled() {
	// Nat enablerer
	go func() {
		for {
			err := enableNat()
			if err == nil {
				logrus.Debug("nat is enabled")
				return
			}
			logrus.Debugf("can not enable nat: %v", err)
			// TODO(cad): employ a exponential back-off approach here
			// instead of sleeping for the constant duration.
			time.Sleep(1 * time.Second)
		}

	}()
}

// enableNat is an idempotent command that ensures nat is enabled for the vpn server.
func enableNat() error {
	if Testing {
		return nil
	}
	// rif := routedInterface("ip", net.FlagUp|net.FlagBroadcast)
	// if rif == nil {
	// 	return fmt.Errorf("can not get routable network interface")
	// }
	rif := getOutboundInterface()
	if rif == nil {
		return fmt.Errorf("can not get default gw interface")
	}

	vpnIfc := vpnInterface()
	if vpnIfc == nil {
		return fmt.Errorf("can not get vpn network interface on the system")
	}

	// Enable ip forwarding.
	emitToFile("/proc/sys/net/ipv4/ip_forward", "1", 0)
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("can not create new iptables object: %v", err)
	}

	server, err := GetServerInstance()
	if err != nil {
		logrus.Errorf("can't get server instance: %v", err)
		return nil
	}

	mask := net.IPMask(net.ParseIP(server.Mask))
	prefix := net.ParseIP(server.Net)
	netw := prefix.Mask(mask).To4()
	netw[3] = byte(1) // Server is always gets xxx.xxx.xxx.1
	ipnet := net.IPNet{IP: netw, Mask: mask}

	// Append iptables nat rules.
	if err := ipt.AppendUnique("nat", "POSTROUTING", "-s", ipnet.String(), "-o", rif.Name, "-j", "MASQUERADE"); err != nil {
		return err
	}

	if err := ipt.AppendUnique("filter", "FORWARD", "-i", rif.Name, "-o", vpnIfc.Name, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "FORWARD", "-i", vpnIfc.Name, "-o", rif.Name, "-j", "ACCEPT"); err != nil {
		return err
	}
	return nil

}

// HostID2IP converts a host id (32-bit unsigned integer) to an IP address.
func HostID2IP(hostid uint32) net.IP {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, hostid)
	return net.IP(ip).To4()
}

// IP2HostID converts an IP address to a host id (32-bit unsigned integer).
func IP2HostID(ip net.IP) uint32 {
	hostid := binary.BigEndian.Uint32(ip.To4())
	return hostid
}

// IncrementIP will return next ip address within the network.
func IncrementIP(ip, mask string) (string, error) {
	if !govalidator.IsIPv4(ip) {
		return "", fmt.Errorf("'ip' is expected to be a valid IPv4 %s", ip)
	}
	if !govalidator.IsIPv4(ip) {
		return "", fmt.Errorf("'mask' is expected to be a valid IPv4 %s", mask)
	}

	ipAddr := net.ParseIP(ip).To4()
	netMask := net.IPMask(net.ParseIP(mask).To4())
	ipNet := net.IPNet{IP: ipAddr, Mask: netMask}
	for i := len(ipAddr) - 1; i >= 0; i-- {
		ipAddr[i]++
		if ip[i] != 0 {
			break
		}
	}
	if !ipNet.Contains(ipAddr) {
		return ip, errors.New("overflowed CIDR while incrementing IP")
	}
	return ipAddr.String(), nil
}
