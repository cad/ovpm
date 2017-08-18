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

var action string

func main() {
	app := cli.NewApp()
	app.Name = "ovpm"
	app.Usage = "OpenVPN Manager"
	app.Version = ovpm.Version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:  "daemon-port",
			Usage: "port number for OVPM daemon to call",
		},
	}
	app.Before = func(c *cli.Context) error {
		logrus.SetLevel(logrus.InfoLevel)
		if c.GlobalBool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil

	}
	app.Commands = []cli.Command{
		{
			Name:  "user",
			Usage: "User Operations",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List VPN users.",
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
						table.SetHeader([]string{"#", "username", "ip", "created at", "valid crt", "no gw"})
						//table.SetBorder(false)
						for i, user := range resp.Users {
							data := []string{fmt.Sprintf("%v", i+1), user.Username, user.IPNet, user.CreatedAt, fmt.Sprintf("%t", user.ServerSerialNumber == server.SerialNumber), fmt.Sprintf("%t", user.NoGW)}
							table.Append(data)
						}
						table.Render()

						return nil
					},
				},
				{
					Name:  "create",
					Usage: "Create a VPN user.",
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
					},
					Action: func(c *cli.Context) error {
						action = "user:create"
						username := c.String("username")
						password := c.String("password")
						noGW := c.Bool("no-gw")

						if username == "" || password == "" {
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)
						}

						//conn := getConn(c.String("port"))
						conn := getConn(c.GlobalString("daemon-port"))
						defer conn.Close()
						userSvc := pb.NewUserServiceClient(conn)

						response, err := userSvc.Create(context.Background(), &pb.UserCreateRequest{Username: username, Password: password, NoGW: noGW})
						if err != nil {
							logrus.Errorf("user can not be created '%s': %v", username, err)
							os.Exit(1)
							return err
						}
						logrus.Infof("user created: %s", response.Users[0].Username)
						return nil
					},
				},
				{
					Name:  "update",
					Usage: "Update a VPN user.",
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
					},
					Action: func(c *cli.Context) error {
						action = "user:update"
						username := c.String("username")
						password := c.String("password")
						nogw := c.Bool("no-gw")
						gw := c.Bool("gw")

						if username == "" {
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)
						}

						if !(password != "" || gw || nogw) {
							fmt.Println("nothing is updated!")
							fmt.Println()
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)
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

						//conn := getConn(c.String("port"))
						conn := getConn(c.GlobalString("daemon-port"))
						defer conn.Close()
						userSvc := pb.NewUserServiceClient(conn)

						response, err := userSvc.Update(context.Background(), &pb.UserUpdateRequest{
							Username: username,
							Password: password,
							Gwpref:   gwPref,
						})

						if err != nil {
							logrus.Errorf("user can not be updated '%s': %v", username, err)
							os.Exit(1)
							return err
						}
						logrus.Infof("user updated: %s", response.Users[0].Username)
						return nil
					},
				},
				{
					Name:  "delete",
					Usage: "Delete a VPN user.",
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
				},
				{
					Name:  "renew",
					Usage: "Renew VPN user certificates.",
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
				},
				{
					Name:  "genconfig",
					Usage: "Generate client config for the user. (.ovpn file)",
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
				},
			},
		},
		{
			Name:  "vpn",
			Usage: "VPN Operations",
			Subcommands: []cli.Command{
				{
					Name:  "status",
					Usage: "Show VPN status.",
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
						table.Append([]string{"Network", res.Net})
						table.Append([]string{"Netmask", res.Mask})
						table.Append([]string{"Created At", res.CreatedAt})
						table.Render()

						return nil
					},
				},
				{
					Name:  "init",
					Usage: "Initialize VPN server.",
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
								if _, err := vpnSvc.Init(context.Background(), &pb.VPNInitRequest{Hostname: hostname, Port: port}); err != nil {
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
				},
			},
		},
	}
	app.Run(os.Args)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
