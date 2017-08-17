package ovpm

import (
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-iptables/iptables"
	"time"
)

// routedInterface returns a network interface that can route IP
// traffic and satisfies flags. It returns nil when an appropriate
// network interface is not found. Network must be "ip", "ip4" or
// "ip6".
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
	mask := net.IPMask(net.ParseIP(_DefaultServerNetMask))
	prefix := net.ParseIP(_DefaultServerNetwork)
	netw := prefix.Mask(mask).To4()
	netw[3] = byte(1) // Server is always gets xxx.xxx.xxx.1
	ipnet := net.IPNet{IP: netw, Mask: mask}

	ifs, err := net.Interfaces()
	if err != nil {
		logrus.Errorf("can not get system network interfaces: %v", err)
		return nil
	}

	for _, ifc := range ifs {
		addrs, err := ifc.Addrs()
		if err != nil {
			logrus.Errorf("can not get interface addresses: %v", err)
			return nil
		}
		for _, addr := range addrs {
			//logrus.Debugf("addr: %s == %s", addr.String(), ipnet.String())
			if addr.String() == ipnet.String() {
				return &ifc
			}
		}
	}
	return nil
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
	rif := routedInterface("ip", net.FlagUp|net.FlagBroadcast)
	if rif == nil {
		return fmt.Errorf("can not get routable network interface")
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

	// Append iptables nat rules.
	ipt.AppendUnique("nat", "POSTROUTING", "-o", rif.Name, "-j", "MASQUERADE")
	// TODO(cad): we should use the interface name that we get when we query the system
	// with the vpn server's internal ip address, instead of default "tun0".
	ipt.AppendUnique("filter", "FORWARD", "-i", rif.Name, "-o", vpnIfc.Name, "-m", "state", "--state", "RELATED, ESTABLISHED", "-j", "ACCEPT")
	ipt.AppendUnique("filter", "FORWARD", "-i", vpnIfc.Name, "-o", rif.Name, "-j", "ACCEPT")
	return nil
}
