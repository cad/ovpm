package main

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/cad/ovpm/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"go.uber.org/thriftrw/ptr"
)

var vpnStatusCommand = cli.Command{
	Name:    "status",
	Usage:   "Show VPN status.",
	Aliases: []string{"s"},
	Action: func(c *cli.Context) error {
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return vpnStatusAction(fmt.Sprintf("grpc://localhost:%d", daemonPort))
	},
}

var vpnInitCommand = cli.Command{
	Name:    "init",
	Usage:   "Initialize VPN server.",
	Aliases: []string{"i"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "hostname, s",
			Usage: "ip address or FQDN of the vpn server",
		},
		cli.StringFlag{
			Name:  "port, p",
			Usage: "port number of the vpn server",
			Value: ovpm.DefaultVPNPort,
		},
		cli.BoolFlag{
			Name:  "tcp, t",
			Usage: "use TCP for vpn protocol, instead of UDP",
		},
		cli.StringFlag{
			Name:  "net, n",
			Usage: "VPN network to give clients IP addresses from, in the CIDR form",
			Value: ovpm.DefaultVPNNetwork,
		},
		cli.StringFlag{
			Name:  "dns, d",
			Usage: fmt.Sprintf("DNS server to push to clients (default: %s)", ovpm.DefaultVPNDNS),
		},
		cli.StringFlag{
			Name:  "keepalive-period",
			Usage: "Ping period to check if the remote peer is alive.",
			Value: ovpm.DefaultKeepalivePeriod,
		},
		cli.StringFlag{
			Name:  "keepalive-timeout",
			Usage: "Ping timeout to assume that remote peer is down.",
			Value: ovpm.DefaultKeepaliveTimeout,
		},
		cli.BoolFlag{
			Name:  "use-lzo, l",
			Usage: "Used to determine whether to use the deprecated lzo compression algorithm to support older clients. (default: false)",
		},
	},
	Action: func(c *cli.Context) error {
		action = "vpn:init"
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate hostname.
		hostname := c.String("hostname")
		if govalidator.IsNull(hostname) || !govalidator.IsHost(hostname) {
			return errors.NotHostname(hostname)
		}

		// Set port number, if provided.
		port := c.String("port")
		if !govalidator.IsNumeric(port) {
			return errors.InvalidPort(port)
		}

		// Set proto if provided.
		proto := pb.VPNProto_UDP
		if c.Bool("tcp") {
			proto = pb.VPNProto_TCP
		}

		// Set ipblock if provided.
		netCIDR := c.String("net")
		if !govalidator.IsCIDR(netCIDR) {
			return errors.NotCIDR(netCIDR)
		}

		// Set DNS if provided.
		dnsAddr := ovpm.DefaultVPNDNS
		if !govalidator.IsIPv4(dnsAddr) {
			return errors.NotIPv4(dnsAddr)
		}

		// Set KeepalivePeriod if provided.
		keepalivePeriod := c.String("keepalive-period")
		if !govalidator.IsNumeric(keepalivePeriod) {
			return errors.NotValidKeepalivePeriod(keepalivePeriod)
		}

		// Set KeepaliveTimeout if provided.
		keepaliveTimeout := c.String("keepalive-timeout")
		if !govalidator.IsNumeric(keepaliveTimeout) {
			return errors.NotValidKeepaliveTimeout(keepaliveTimeout)
		}

		useLZO := c.Bool("use-lzo")

		// Ask for confirmation from the user about the destructive
		// changes that are about to happen.
		var uiConfirmed bool
		{
			var response string
			for {
				fmt.Println("This operation will cause invalidation of existing user certificates.")
				fmt.Println("After this opeartion, new client config files (.ovpn) should be generated for each existing user.")
				fmt.Println()
				fmt.Println("Are you sure ? (y/N)")
				_, err := fmt.Scanln(&response)
				if err != nil {
					logrus.Fatal(err)
					exit(1)
					return err
				}
				okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
				nokayResponses := []string{"n", "N", "no", "No", "NO"}
				if stringInSlice(response, okayResponses) {
					uiConfirmed = true
					break
				}
				if stringInSlice(response, nokayResponses) {
					uiConfirmed = false
					break
				}
			}
		}

		// Did user confirm the destructive changes?
		if !uiConfirmed {
			return errors.Unconfirmed("user decided to cancel")
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		err := vpnInitAction(vpnInitParams{
			rpcServURLStr:    fmt.Sprintf("grpc://localhost:%d", daemonPort),
			hostname:         hostname,
			port:             port,
			proto:            proto,
			netCIDR:          netCIDR,
			dnsAddr:          c.String("dns"),
			keepalivePeriod:  keepalivePeriod,
			keepaliveTimeout: keepaliveTimeout,
			useLZO:           useLZO,
		})
		if err != nil {
			e, ok := err.(errors.Error)
			if ok {
				switch e.Code {
				case errors.ErrNotHostname:
					fmt.Printf("--hostname option requires a valid hostname: '%s' is not a hostname", c.String("hostname"))
					exit(1)
					return e
				}
			}
			return err
		}
		return nil
	},
}

var vpnUpdateCommand = cli.Command{
	Name:    "update",
	Usage:   "Update VPN server.",
	Aliases: []string{"u"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: fmt.Sprintf("VPN network to give clients IP addresses from, in the CIDR form (default: %s)", ovpm.DefaultVPNNetwork),
		},
		cli.StringFlag{
			Name:  "dns, d",
			Usage: fmt.Sprintf("DNS server to push to clients (default: %s)", ovpm.DefaultVPNDNS),
		},
		cli.BoolFlag{
			Name:  "enable-use-lzo",
			Usage: fmt.Sprintf("Enable use of the deprecated lzo compression algorithm to support older clients."),
		},
		cli.BoolFlag{
			Name:  "disable-use-lzo",
			Usage: fmt.Sprintf("Disable use of the deprecated lzo compression algorithm to support older clients."),
		},
	},
	Action: func(c *cli.Context) error {
		action = "vpn:update"
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		var netCIDR *string
		if net := c.String("net"); !govalidator.IsNull(net) {
			netCIDR = &net
		}

		var dnsAddr *string
		if dns := c.String("dns"); !govalidator.IsNull(dns) {
			dnsAddr = &dns
		}

		var useLzo *bool
		if c.Bool("enable-use-lzo") && c.Bool("disable-use-lzo") {
			e := fmt.Errorf("can not use --enable-use-lzo and --disable-use-lzo together")
			fmt.Println(e.Error())
			exit(1)
			return e
		}
		if enableLzo := c.Bool("enable-use-lzo"); enableLzo {
			useLzo = ptr.Bool(true)
		}
		if disableLzo := c.Bool("disable-use-lzo"); disableLzo {
			useLzo = ptr.Bool(false)
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return vpnUpdateAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), netCIDR, dnsAddr, useLzo)
	},
}

var vpnRestartCommand = cli.Command{
	Name:    "restart",
	Usage:   "Restart VPN server.",
	Aliases: []string{"r"},
	Action: func(c *cli.Context) error {
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return vpnRestartAction(fmt.Sprintf("grpc://localhost:%d", daemonPort))
	},
}

func init() {
	app.Commands = append(app.Commands,
		cli.Command{
			Name:    "vpn",
			Usage:   "VPN Operations",
			Aliases: []string{"v"},
			Subcommands: []cli.Command{
				vpnStatusCommand,
				vpnInitCommand,
				vpnUpdateCommand,
				vpnRestartCommand,
			},
		},
	)
}
