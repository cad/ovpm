package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm/pb"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var netDefineCommand = cli.Command{
	Name:  "define",
	Usage: "Define a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cidr, c",
			Usage: "CIDR of the network",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of the network",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:create"
		name := c.String("name")
		cidr := c.String("cidr")

		if name == "" || cidr == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		netSvc := pb.NewNetworkServiceClient(conn)

		response, err := netSvc.Create(context.Background(), &pb.NetworkCreateRequest{Name: name, CIDR: cidr})
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
	Name:  "list",
	Usage: "List network definitions.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cidr, c",
			Usage: "CIDR of the network",
		},
	},
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
		table.SetHeader([]string{"#", "name", "cidr", "created at"})
		//table.SetBorder(false)
		for i, network := range resp.Networks {
			data := []string{fmt.Sprintf("%v", i+1), network.Name, network.CIDR, network.CreatedAt}
			table.Append(data)
		}
		table.Render()

		return nil
	},
}

var netUndefineCommand = cli.Command{
	Name:  "undefine",
	Usage: "Undefine an existing network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of the network",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:delete"
		name := c.String("name")

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
