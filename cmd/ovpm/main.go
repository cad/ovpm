//go:generate go-bindata template/

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
		logrus.SetLevel(logrus.WarnLevel)
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
						table.SetHeader([]string{"#", "username", "created at", "valid crt"})
						//table.SetBorder(false)
						for i, user := range resp.Users {
							data := []string{fmt.Sprintf("%v", i+1), user.Username, user.CreatedAt, fmt.Sprintf("%t", user.ServerSerialNumber == server.SerialNumber)}
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
					},
					Action: func(c *cli.Context) error {
						action = "user:create"
						username := c.String("username")
						password := c.String("password")

						if username == "" || password == "" {
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)
						}

						//conn := getConn(c.String("port"))
						conn := getConn(c.GlobalString("daemon-port"))
						defer conn.Close()
						userSvc := pb.NewUserServiceClient(conn)

						response, err := userSvc.Create(context.Background(), &pb.UserCreateRequest{Username: username, Password: password})
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
						logrus.Infof("user certs renewed: '%s'", username)
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
						emitToFile(output, res.ClientConfig, 0)

						if err != nil {
							logrus.Errorf("user config can not be exported %s: %v", username, err)
							return err
						}
						fmt.Printf("exported to %s", output)
						return nil
					},
				},
				// {
				// 	Name:  "lock",
				// 	Usage: "Lock VPN user",
				// 	Action: func(c *cli.Context) error {
				// 		return nil
				// 	},
				// },
				// {
				// 	Name:  "unlock",
				// 	Usage: "Unlock VPN user",
				// 	Action: func(c *cli.Context) error {
				// 		return nil
				// 	},
				// },
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

								break
							} else if stringInSlice(response, nokayResponses) {
								return fmt.Errorf("user decided to cancel")
							}
						}

						return nil
					},
				},
				{
					Name:  "apply",
					Usage: "Apply pending changes.",
					Action: func(c *cli.Context) error {
						action = "apply"

						conn := getConn(c.GlobalString("daemon-port"))
						defer conn.Close()
						vpnSvc := pb.NewVPNServiceClient(conn)

						if _, err := vpnSvc.Apply(context.Background(), &pb.VPNApplyRequest{}); err != nil {
							logrus.Errorf("can not apply configuration: %v", err)
							os.Exit(1)
							return err
						}
						logrus.Info("changes applied")
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
