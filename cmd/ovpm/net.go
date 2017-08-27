package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/pb"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var netDefineCommand = cli.Command{
	Name:    "def",
	Aliases: []string{"d"},
	Usage:   "Define a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cidr, c",
			Usage: "CIDR of the network",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of the network",
		},
		cli.StringFlag{
			Name:  "type, t",
			Usage: "type of the network (see $ovpm net types)",
		},
		cli.StringFlag{
			Name:  "via, v",
			Usage: "if network type is route, via represents route's gateway",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:create"
		name := c.String("name")
		cidr := c.String("cidr")
		typ := c.String("type")
		via := c.String("via")

		if name == "" || cidr == "" || typ == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		switch ovpm.NetworkTypeFromString(typ) {
		case ovpm.ROUTE:
			if via != "" && !govalidator.IsCIDR(via) {
				fmt.Printf("validation error: `%s` must be a network in the CIDR form", via)
				fmt.Println()
				fmt.Println(cli.ShowSubcommandHelp(c))
				os.Exit(1)
			} else {
				via = ""
			}
		case ovpm.SERVERNET:
			if via != "" {
				fmt.Println("--via flag can only be used with --type ROUTE")
				fmt.Println()
				fmt.Println(cli.ShowSubcommandHelp(c))
				os.Exit(1)
			}
		default: // Means UNDEFINEDNET
			fmt.Printf("undefined network type %s", typ)
			fmt.Println()
			fmt.Println("Network Types:")
			fmt.Println("    ", ovpm.GetAllNetworkTypes())
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		response, err := netSvc.Create(context.Background(), &pb.NetworkCreateRequest{Name: name, CIDR: cidr, Type: typ, Via: via})
		if err != nil {
			logrus.Errorf("network can not be created '%s': %v", name, err)
			os.Exit(1)
			return err
		}
		logrus.Infof("network created: %s (%s)", response.Network.Name, response.Network.CIDR)
		return nil
	},
}

var netListCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "List defined networks.",
	Action: func(c *cli.Context) error {
		action = "net:list"
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		resp, err := netSvc.List(context.Background(), &pb.NetworkListRequest{})
		if err != nil {
			logrus.Errorf("networks can not be fetched: %v", err)
			os.Exit(1)
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "name", "cidr", "type", "assoc", "created at"})
		//table.SetBorder(false)
		for i, network := range resp.Networks {
			// Create associated user list for this network.
			var usernameList string
			assocUsers, err := netSvc.GetAssociatedUsers(context.Background(), &pb.NetworkGetAssociatedUsersRequest{Name: network.Name})
			if err != nil {
				logrus.Errorf("assoc users can not be fetched: %v", err)
				os.Exit(1)
				return err
			}

			usernames := assocUsers.Usernames
			count := len(usernames)
			for i, uname := range usernames {
				if i+1 == count {
					usernameList = usernameList + fmt.Sprintf("%s", uname)
				} else {
					usernameList = usernameList + fmt.Sprintf("%s, ", uname)
				}
			}
			var cidr = network.CIDR
			var via = network.Via
			if via == "" {
				via = "vpn-server"
			}
			if ovpm.NetworkTypeFromString(network.Type) == ovpm.ROUTE {
				cidr = fmt.Sprintf("%s via %s", network.CIDR, via)
			}
			data := []string{fmt.Sprintf("%v", i+1), network.Name, cidr, network.Type, usernameList, network.CreatedAt}
			table.Append(data)
		}
		table.Render()

		return nil
	},
}

var netTypesCommand = cli.Command{
	Name:    "types",
	Aliases: []string{"t"},
	Usage:   "Show available network types.",
	Action: func(c *cli.Context) error {
		action = "net:types"
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		resp, err := netSvc.GetAllTypes(context.Background(), &pb.NetworkGetAllTypesRequest{})
		if err != nil {
			logrus.Errorf("networks can not be fetched: %v", err)
			os.Exit(1)
			return err
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "net type", "desc"})
		//table.SetBorder(false)
		for i, ntype := range resp.Types {
			data := []string{fmt.Sprintf("%v", i+1), ntype.Type, ntype.Description}
			table.Append(data)
		}
		table.Render()

		return nil
	},
}

var netUndefineCommand = cli.Command{
	Name:    "undef",
	Aliases: []string{"u"},
	Usage:   "Undefine an existing network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:delete"
		name := c.String("net")

		if name == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		resp, err := netSvc.Delete(context.Background(), &pb.NetworkDeleteRequest{Name: name})
		if err != nil {
			logrus.Errorf("networks can not be deleted: %v", err)
			os.Exit(1)
			return err
		}
		logrus.Infof("network deleted: %s (%s)", resp.Network.Name, resp.Network.CIDR)

		return nil
	},
}

var netAssociateCommand = cli.Command{
	Name:    "assoc",
	Aliases: []string{"a"},
	Usage:   "Associate a user with a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},

		cli.StringFlag{
			Name:  "user, u",
			Usage: "name of the user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:associate"
		netName := c.String("net")
		userName := c.String("user")

		if netName == "" || userName == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		_, err := netSvc.Associate(context.Background(), &pb.NetworkAssociateRequest{Name: netName, Username: userName})
		if err != nil {
			logrus.Errorf("networks can not be associated: %v", err)
			os.Exit(1)
			return err
		}
		logrus.Infof("network associated: user:%s <-> network:%s", userName, netName)

		return nil
	},
}

var netDissociateCommand = cli.Command{
	Name:    "dissoc",
	Aliases: []string{"d"},
	Usage:   "Dissociate a user from a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},

		cli.StringFlag{
			Name:  "user, u",
			Usage: "name of the user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:dissociate"
		netName := c.String("net")
		userName := c.String("user")

		if netName == "" || userName == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		_, err := netSvc.Dissociate(context.Background(), &pb.NetworkDissociateRequest{Name: netName, Username: userName})
		if err != nil {
			logrus.Errorf("networks can not be dissociated: %v", err)
			os.Exit(1)
			return err
		}
		logrus.Infof("network dissociated: user:%s <-> network:%s", userName, netName)

		return nil
	},
}
