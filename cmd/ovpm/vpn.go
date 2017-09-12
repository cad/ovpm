package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var vpnStatusCommand = cli.Command{
	Name:    "status",
	Usage:   "Show VPN status.",
	Aliases: []string{"s"},
	Action: func(c *cli.Context) error {
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		vpnSvc := pb.NewVPNServiceClient(conn)

		res, err := vpnSvc.Status(context.Background(), &pb.VPNStatusRequest{})
		if err != nil {
			os.Exit(1)
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"attribute", "value"})
		table.Append([]string{"Name", res.Name})
		table.Append([]string{"Hostname", res.Hostname})
		table.Append([]string{"Port", res.Port})
		table.Append([]string{"Proto", res.Proto})
		table.Append([]string{"Network", res.Net})
		table.Append([]string{"Netmask", res.Mask})
		table.Append([]string{"Created At", res.CreatedAt})
		table.Append([]string{"DNS", res.Dns})
		table.Render()

		return nil
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
			Usage: fmt.Sprintf("VPN network to give clients IP addresses from, in the CIDR form (default: %s)", ovpm.DefaultVPNNetwork),
		},
		cli.StringFlag{
			Name:  "dns, d",
			Usage: fmt.Sprintf("DNS server to push to clients (default: %s)", ovpm.DefaultVPNDNS),
		},
	},
	Action: func(c *cli.Context) error {
		action = "vpn:init"
		hostname := c.String("hostname")
		if hostname == "" {
			logrus.Errorf("'hostname' is required")
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)

		}

		port := c.String("port")
		if port == "" {
			port = ovpm.DefaultVPNPort
		}

		tcp := c.Bool("tcp")

		proto := pb.VPNProto_UDP
		if tcp {
			proto = pb.VPNProto_TCP
		}

		ipblock := c.String("net")
		if ipblock != "" && !govalidator.IsCIDR(ipblock) {
			fmt.Println("--net takes an ip network in the CIDR form. e.g. 10.9.0.0/24")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		dns := c.String("dns")
		if dns != "" && !govalidator.IsIPv4(dns) {
			fmt.Println("--dns takes an IPv4 address. e.g. 8.8.8.8")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		vpnSvc := pb.NewVPNServiceClient(conn)

		var response string
		for {
			fmt.Println("This operation will cause invalidation of existing user certificates.")
			fmt.Println("After this opeartion, new client config files (.ovpn) should be generated for each existing user.")
			fmt.Println()
			fmt.Println("Are you sure ? (y/N)")
			_, err := fmt.Scanln(&response)
			if err != nil {
				logrus.Fatal(err)
				os.Exit(1)
				return err
			}
			okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
			nokayResponses := []string{"n", "N", "no", "No", "NO"}
			if stringInSlice(response, okayResponses) {
				if _, err := vpnSvc.Init(context.Background(), &pb.VPNInitRequest{Hostname: hostname, Port: port, ProtoPref: proto, IpBlock: ipblock, Dns: dns}); err != nil {
					logrus.Errorf("server can not be initialized: %v", err)
					os.Exit(1)
					return err
				}
				logrus.Info("ovpm server initialized")
				break
			} else if stringInSlice(response, nokayResponses) {
				return fmt.Errorf("user decided to cancel")
			}
		}

		return nil
	},
}

var vpnUpdateCommand = cli.Command{
	Name:    "update",
	Usage:   "Update VPN server.",
	Aliases: []string{"i"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: fmt.Sprintf("VPN network to give clients IP addresses from, in the CIDR form (default: %s)", ovpm.DefaultVPNNetwork),
		},
		cli.StringFlag{
			Name:  "dns, d",
			Usage: fmt.Sprintf("DNS server to push to clients (default: %s)", ovpm.DefaultVPNDNS),
		},
	},
	Action: func(c *cli.Context) error {
		action = "vpn:update"

		ipblock := c.String("net")
		if ipblock != "" && !govalidator.IsCIDR(ipblock) {
			fmt.Println("--net takes an ip network in the CIDR form. e.g. 10.9.0.0/24")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		if ipblock != "" {
			var response string
			for {
				fmt.Println("If you proceed, you will loose all your static ip definitions.")
				fmt.Println("Any user that is defined to have a static ip will be set to be dynamic again.")
				fmt.Println()
				fmt.Println("Are you sure ? (y/N)")
				_, err := fmt.Scanln(&response)
				if err != nil {
					logrus.Fatal(err)
					os.Exit(1)
					return err
				}
				okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
				nokayResponses := []string{"n", "N", "no", "No", "NO"}
				if stringInSlice(response, okayResponses) {
					break
				} else if stringInSlice(response, nokayResponses) {
					return fmt.Errorf("user decided to cancel")
				}
			}

		}

		dns := c.String("dns")
		if dns != "" && !govalidator.IsIPv4(dns) {
			fmt.Println("--dns takes an IPv4 address. e.g. 8.8.8.8")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		if !(ipblock != "" || dns != "") {
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		vpnSvc := pb.NewVPNServiceClient(conn)

		if _, err := vpnSvc.Update(context.Background(), &pb.VPNUpdateRequest{IpBlock: ipblock, Dns: dns}); err != nil {
			logrus.Errorf("server can not be updated: %v", err)
			os.Exit(1)
			return err
		}
		logrus.Info("ovpm server updated")
		return nil
	},
}
