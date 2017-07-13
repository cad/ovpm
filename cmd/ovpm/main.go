//go:generate go-bindata template/
package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"os"
	"time"
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
						server, err := ovpm.GetServerInstance()
						if err != nil {
							os.Exit(1)
							return err
						}
						users, err := ovpm.GetAllUsers()
						if err != nil {
							logrus.Errorf("users can not be fetched: %v", err)
							os.Exit(1)
							return err
						}
						table := tablewriter.NewWriter(os.Stdout)
						table.SetHeader([]string{"#", "username", "created at", "valid crt"})
						//table.SetBorder(false)
						for i, user := range users {
							data := []string{fmt.Sprintf("%v", i+1), user.Username, user.CreatedAt.Format(time.UnixDate), fmt.Sprintf("%t", server.CheckSerial(user.ServerSerialNumber))}
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
						user, err := ovpm.CreateUser(username, password)
						if err != nil {
							logrus.Errorf("user can not be created '%s': %v", username, err)
							os.Exit(1)
							return err
						}
						logrus.Infof("user created: %s", user.Username)
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
						err := ovpm.DeleteUser(username)
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
						err := ovpm.SignUser(username)
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
						err := ovpm.DumpUserOVPNConf(username, output)
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
						server, err := ovpm.GetServerInstance()
						if err != nil {
							os.Exit(1)
							return err
						}

						table := tablewriter.NewWriter(os.Stdout)
						table.SetHeader([]string{"attribute", "value"})
						table.Append([]string{"Name", server.Name})
						table.Append([]string{"Hostname", server.Hostname})
						table.Append([]string{"Port", server.Port})
						table.Append([]string{"Network", server.Net})
						table.Append([]string{"Netmask", server.Mask})
						table.Append([]string{"Created At", server.CreatedAt.Format(time.UnixDate)})
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
						if c.String("hostname") == "" {
							logrus.Errorf("'hostname' is needed")
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)

						}

						if ovpm.CheckBootstrapped() {
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
									if err := ovpm.DeleteServer("default"); err != nil {
										logrus.Errorf("server can not be deleted: %v", err)
										os.Exit(1)
										return err
									}

									break
								} else if stringInSlice(response, nokayResponses) {
									return fmt.Errorf("user decided to cancel")
								}
							}

						}
						if err := ovpm.CreateServer("default", c.String("hostname"), c.String("port")); err != nil {
							logrus.Errorf("server can not be created: %v", err)
							fmt.Println(cli.ShowSubcommandHelp(c))
							os.Exit(1)
						}

						return nil
					},
				},
				{
					Name:  "apply",
					Usage: "Apply pending changes.",
					Action: func(c *cli.Context) error {
						action = "apply"
						if err := ovpm.Emit(); err != nil {
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
	ovpm.CloseDB()
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
