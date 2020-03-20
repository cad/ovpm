package main

import (
	"os"

	"github.com/cad/ovpm"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var app = cli.NewApp()

func main() {
	app.Run(os.Args)
}

func init() {
	app.Name = "ovpm"
	app.Usage = "OpenVPN Manager"
	app.Version = ovpm.Version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose output",
		},
		cli.IntFlag{
			Name:  "daemon-port",
			Usage: "port number for OVPM daemon to call",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "just validate command flags; not make any calls to the daemon behind",
		},
	}
	app.Before = func(c *cli.Context) error {
		logrus.SetLevel(logrus.InfoLevel)
		if c.GlobalBool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
}
