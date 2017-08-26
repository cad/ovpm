package ovpm

import (
	"encoding/binary"
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
	Type   NetworkType
	String string
}{
	{UNDEFINEDNET, "UNDEFINEDNET"},
	{SERVERNET, "SERVERNET"},
	{ROUTE, "ROUTE"},
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

// DBNetwork is database model for external networks on the VPN server.
type DBNetwork struct {
	gorm.Model
	ServerID uint
	Server   DBServer

	Name  string `gorm:"unique_index"`
	CIDR  string
	Type  NetworkType
	Users []*DBUser `gorm:"many2many:network_users;"`
}

// GetNetwork returns a network specified by its name.
func GetNetwork(name string) (*DBNetwork, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(name) {
		return nil, fmt.Errorf("validation error: %s can not be null", name)
	}
	if !govalidator.IsAlphanumeric(name) {
		return nil, fmt.Errorf("validation error: `%s` can only contain letters and numbers", name)
	}

	var network DBNetwork
	db.Preload("Users").Where(&DBNetwork{Name: name}).First(&network)

	if db.NewRecord(&network) {
		return nil, fmt.Errorf("network not found %s", name)
	}

	return &network, nil
}

// GetAllNetworks returns all networks defined in the system.
func GetAllNetworks() []*DBNetwork {
	var networks []*DBNetwork
	db.Preload("Users").Find(&networks)

	return networks
}

// CreateNewNetwork creates a new network definition in the system.
func CreateNewNetwork(name, cidr string, nettype NetworkType) (*DBNetwork, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("you first need to create server")
	}
	// Validate user input.
	if govalidator.IsNull(name) {
		return nil, fmt.Errorf("validation error: %s can not be null", name)
	}
	if !govalidator.IsAlphanumeric(name) {
		return nil, fmt.Errorf("validation error: `%s` can only contain letters and numbers", name)
	}

	if !govalidator.IsCIDR(cidr) {
		return nil, fmt.Errorf("validation error: `%s` must be a network in the CIDR form", cidr)
	}

	if nettype == UNDEFINEDNET {
		return nil, fmt.Errorf("validation error: `%s` must be a valid network type", nettype)
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("can not parse CIDR %s: %v", cidr, err)
	}

	network := DBNetwork{
		Name:  name,
		CIDR:  ipnet.String(),
		Type:  nettype,
		Users: []*DBUser{},
	}
	db.Save(&network)

	if db.NewRecord(&network) {
		return nil, fmt.Errorf("can not create network in the db")
	}
	Emit()
	logrus.Infof("network defined: %s (%s)", network.Name, network.CIDR)
	return &network, nil

}

// Delete deletes a network definition in the system.
func (n *DBNetwork) Delete() error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}

	db.Unscoped().Delete(n)
	Emit()
	logrus.Infof("network deleted: %s", n.Name)
	return nil
}

// Associate allows the given user access to this network.
func (n *DBNetwork) Associate(username string) error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}
	user, err := GetUser(username)
	if err != nil {
		return fmt.Errorf("user can not be fetched: %v", err)
	}

	var users []DBUser
	userAssoc := db.Model(&n).Association("Users")
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

	userAssoc.Append(user)
	if userAssoc.Error != nil {
		return fmt.Errorf("association failed: %v", userAssoc.Error)
	}
	Emit()
	logrus.Infof("user '%s' is associated with the network '%s'", user.GetUsername(), n.Name)
	return nil
}

// Dissociate breaks up the given users association to the said network.
func (n *DBNetwork) Dissociate(username string) error {
	if !IsInitialized() {
		return fmt.Errorf("you first need to create server")
	}

	user, err := GetUser(username)
	if err != nil {
		return fmt.Errorf("user can not be fetched: %v", err)
	}

	var users []DBUser
	userAssoc := db.Model(&n).Association("Users")
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

	userAssoc.Delete(user)
	if userAssoc.Error != nil {
		return fmt.Errorf("disassociation failed: %v", userAssoc.Error)
	}
	Emit()
	logrus.Infof("user '%s' is dissociated with the network '%s'", user.GetUsername(), n.Name)
	return nil
}

// GetName returns network's name.
func (n *DBNetwork) GetName() string {
	return n.Name
}

// GetCIDR returns network's CIDR.
func (n *DBNetwork) GetCIDR() string {
	return n.CIDR
}

// GetCreatedAt returns network's name.
func (n *DBNetwork) GetCreatedAt() string {
	return n.CreatedAt.Format(time.UnixDate)
}

// GetType returns network's network type.
func (n *DBNetwork) GetType() NetworkType {
	return NetworkType(n.Type)
}

// GetAssociatedUsers returns network's associated users.
func (n *DBNetwork) GetAssociatedUsers() []*DBUser {
	return n.Users
}

// GetAssociatedUsernames returns network's associated user names.
func (n *DBNetwork) GetAssociatedUsernames() []string {
	var usernames []string

	for _, user := range n.GetAssociatedUsers() {
		usernames = append(usernames, user.Username)
	}
	return usernames
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
	return net.IP(ip)
}

//IP2HostID converts an IP address to a host id (32-bit unsigned integer).
func IP2HostID(ip net.IP) uint32 {
	hostid := binary.BigEndian.Uint32(ip)
	return hostid
}
