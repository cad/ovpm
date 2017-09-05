package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var userListCommand = cli.Command{
	Name:    "list",
	Usage:   "List VPN users.",
	Aliases: []string{"l"},
	Action: func(c *cli.Context) error {
		action = "user:list"
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)
		vpnSvc := pb.NewVPNServiceClient(conn)

		server, err := vpnSvc.Status(context.Background(), &pb.VPNStatusRequest{})
		if err != nil {
			logrus.Errorf("can not get server status: %v", err)
			os.Exit(1)
			return err
		}

		resp, err := userSvc.List(context.Background(), &pb.UserListRequest{})
		if err != nil {
			logrus.Errorf("users can not be fetched: %v", err)
			os.Exit(1)
			return err
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "username", "ip", "created at", "valid crt", "push gw"})
		//table.SetBorder(false)
		for i, user := range resp.Users {
			static := ""
			if user.HostID != 0 {
				static = "s"
			}
			username := user.Username
			if user.IsAdmin {
				username = fmt.Sprintf("%s *", username)
			}
			data := []string{fmt.Sprintf("%v", i+1), username, fmt.Sprintf("%s %s", user.IPNet, static), user.CreatedAt, fmt.Sprintf("%t", user.ServerSerialNumber == server.SerialNumber), fmt.Sprintf("%t", !user.NoGW)}
			table.Append(data)
		}
		table.Render()

		return nil
	},
}

var userCreateCommand = cli.Command{
	Name:    "create",
	Usage:   "Create a VPN user.",
	Aliases: []string{"c"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "username, u",
			Usage: "username for the vpn user",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "password for the vpn user",
		},
		cli.BoolFlag{
			Name:  "no-gw",
			Usage: "don't push vpn server as default gateway for this user",
		},
		cli.StringFlag{
			Name:  "static",
			Usage: "ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "admin, a",
			Usage: "this user has admin rights",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:create"
		username := c.String("username")
		password := c.String("password")
		noGW := c.Bool("no-gw")
		static := c.String("static")
		admin := c.Bool("admin")

		if username == "" || password == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}
		if static != "" && !govalidator.IsIPv4(static) {
			fmt.Println("--static flag takes a valid ipv4 address")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}
		var hostid uint32
		if static != "" {
			h := ovpm.IP2HostID(net.ParseIP(static).To4())
			if h == 0 {
				fmt.Printf("can not parse %s as IPv4", static)
				fmt.Println()
				fmt.Println(cli.ShowSubcommandHelp(c))
				os.Exit(1)
			}

			hostid = h
		}

		//conn := getConn(c.String("port"))
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)

		response, err := userSvc.Create(context.Background(),
			&pb.UserCreateRequest{Username: username, Password: password, NoGW: noGW, HostID: hostid, IsAdmin: admin},
		)
		if err != nil {
			logrus.Errorf("user can not be created '%s': %v", username, err)
			os.Exit(1)
			return err
		}
		logrus.Infof("user created: %s", response.Users[0].Username)
		return nil
	},
}

var userUpdateCommand = cli.Command{
	Name:    "update",
	Usage:   "Update a VPN user.",
	Aliases: []string{"u"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "username, u",
			Usage: "username of the vpn user to update",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "new password for the vpn user",
		},
		cli.BoolFlag{
			Name:  "no-gw",
			Usage: "don't push vpn server as default gateway for this user",
		},
		cli.BoolFlag{
			Name:  "gw",
			Usage: "push vpn server as default gateway for this user",
		},
		cli.StringFlag{
			Name:  "static",
			Usage: "ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "no-static",
			Usage: "do not set static ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "admin",
			Usage: "this user has admin rights",
		},
		cli.BoolFlag{
			Name:  "no-admin",
			Usage: "this user has no admin rights",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:update"
		username := c.String("username")
		password := c.String("password")
		nogw := c.Bool("no-gw")
		gw := c.Bool("gw")
		static := c.String("static")
		noStatic := c.Bool("no-static")
		admin := c.Bool("admin")
		noAdmin := c.Bool("no-admin")

		if username == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		// Check whether if all flags are are empty.
		if !(password != "" || gw || nogw || static != "" || noStatic || admin || noAdmin) {
			fmt.Println("nothing is updated!")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		// Given that static is set, check whether it's IPv4.
		if static != "" && !govalidator.IsIPv4(static) {
			fmt.Println("--static flag takes a valid ipv4 address")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}
		var staticPref pb.UserUpdateRequest_StaticPref
		staticPref = pb.UserUpdateRequest_NOPREFSTATIC
		var hostid uint32

		switch {
		case static != "" && !noStatic:
			// means static is set.
			if static != "" {
				h := ovpm.IP2HostID(net.ParseIP(static).To4())
				if h == 0 {
					fmt.Printf("can't parse %s as IPv4", static)
					fmt.Println()
					fmt.Println(cli.ShowSubcommandHelp(c))
					os.Exit(1)
				}

				hostid = h
			}
			staticPref = pb.UserUpdateRequest_STATIC

		case static == "" && noStatic:
			// means no-static
			hostid = 0
			staticPref = pb.UserUpdateRequest_NOSTATIC
		case static != "" && noStatic:
			// means invalid
			fmt.Println("--static flag and --no-static flag cannot be used together")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		case static == "" && !noStatic:
		default:
			// means no pref
			staticPref = pb.UserUpdateRequest_NOPREFSTATIC
			hostid = 0
		}

		var gwPref pb.UserUpdateRequest_GWPref

		switch {
		case gw && !nogw:
			gwPref = pb.UserUpdateRequest_GW
		case !gw && nogw:
			gwPref = pb.UserUpdateRequest_NOGW
		case gw && nogw:
			// Ambigius.
			fmt.Println("you can't use --gw together with --no-gw")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		default:
			gwPref = pb.UserUpdateRequest_NOPREF

		}

		var adminPref pb.UserUpdateRequest_AdminPref

		switch {
		case admin && !noAdmin:
			adminPref = pb.UserUpdateRequest_ADMIN
		case !admin && noAdmin:
			adminPref = pb.UserUpdateRequest_NOADMIN
		case !admin && !noAdmin:
			adminPref = pb.UserUpdateRequest_NOPREFADMIN
		case admin && noAdmin:
			// Ambigius.
			fmt.Println("you can't use --admin together with --no-admin")
			fmt.Println()
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		//conn := getConn(c.String("port"))
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)

		response, err := userSvc.Update(context.Background(), &pb.UserUpdateRequest{
			Username:   username,
			Password:   password,
			Gwpref:     gwPref,
			HostID:     hostid,
			Staticpref: staticPref,
			Adminpref:  adminPref,
		})

		if err != nil {
			logrus.Errorf("user can not be updated '%s': %v", username, err)
			os.Exit(1)
			return err
		}
		logrus.Infof("user updated: %s", response.Users[0].Username)
		return nil
	},
}

var userDeleteCommand = cli.Command{
	Name:    "delete",
	Usage:   "Delete a VPN user.",
	Aliases: []string{"d"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "username of the vpn user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:delete"
		username := c.String("user")

		if username == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		//conn := getConn(c.String("port"))
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)

		_, err := userSvc.Delete(context.Background(), &pb.UserDeleteRequest{Username: username})
		if err != nil {
			logrus.Errorf("user can not be deleted '%s': %v", username, err)
			os.Exit(1)
			return err
		}
		logrus.Infof("user deleted: %s", username)
		return nil
	},
}

var userRenewCommand = cli.Command{
	Name:    "renew",
	Usage:   "Renew VPN user certificates.",
	Aliases: []string{"r"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "username of the vpn user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:renew"
		username := c.String("user")

		if username == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}

		//conn := getConn(c.String("port"))
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)
		pb.NewVPNServiceClient(conn)

		_, err := userSvc.Renew(context.Background(), &pb.UserRenewRequest{Username: username})
		if err != nil {
			logrus.Errorf("can't renew user cert '%s': %v", username, err)
			os.Exit(1)
			return err
		}
		logrus.Infof("user cert renewed: '%s'", username)
		return nil
	},
}

var userGenconfigCommand = cli.Command{
	Name:    "genconfig",
	Usage:   "Generate client config for the user. (.ovpn file)",
	Aliases: []string{"g"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "username of the vpn user",
		},
		cli.StringFlag{
			Name:  "out, o",
			Usage: ".ovpn file output path",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:export-config"
		username := c.String("user")
		output := c.String("out")

		if username == "" {
			fmt.Println(cli.ShowSubcommandHelp(c))
			os.Exit(1)
		}
		if output == "" {
			output = username + ".ovpn"
		}

		//conn := getConn(c.String("port"))
		conn := getConn(c.GlobalString("daemon-port"))
		defer conn.Close()
		userSvc := pb.NewUserServiceClient(conn)
		pb.NewVPNServiceClient(conn)

		res, err := userSvc.GenConfig(context.Background(), &pb.UserGenConfigRequest{Username: username})
		if err != nil {
			logrus.Errorf("user config can not be exported %s: %v", username, err)
			return err
		}
		emitToFile(output, res.ClientConfig, 0)
		logrus.Infof("exported to %s", output)
		return nil
	},
}
