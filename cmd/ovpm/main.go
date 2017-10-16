package main

import (
	"flag"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/urfave/cli"
)

var action string

func NewApp() *cli.App {
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
			Name:    "user",
			Usage:   "User Operations",
			Aliases: []string{"u"},
			Subcommands: []cli.Command{
				userListCommand,
				userCreateCommand,
				userUpdateCommand,
				userDeleteCommand,
				userRenewCommand,
				userGenconfigCommand,
			},
		},
		{
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
		{
			Name:    "net",
			Usage:   "Network Operations",
			Aliases: []string{"n"},
			Subcommands: []cli.Command{
				netListCommand,
				netTypesCommand,
				netDefineCommand,
				netUndefineCommand,
				netAssociateCommand,
				netDissociateCommand,
			},
		},
	}
	return app
}
func main() {
	app := NewApp()
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

func exit(status int) {
	if flag.Lookup("test.v") == nil {
		os.Exit(status)
	} else {

	}
}
