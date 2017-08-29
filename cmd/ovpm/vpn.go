package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/pb"
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
	},
	Action: func(c *cli.Context) error {
		action = "vpn:init"
		hostname := c.String("hostname")
		if hostname == "" {
			logrus.Errorf("'hostname' is needed")
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)

		}

		port := c.String("port")
		if port == "" {
			port = ovpm.DefaultVPNPort
		}

		tcp := c.Bool("tcp")

		var proto pb.VPNProto

		switch tcp {
		case true:
			proto = pb.VPNProto_TCP
		default:
			proto = pb.VPNProto_UDP
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
				if _, err := vpnSvc.Init(context.Background(), &pb.VPNInitRequest{Hostname: hostname, Port: port, Protopref: proto}); err != nil {
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
