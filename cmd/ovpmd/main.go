//go:generate go-bindata template/

package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api"
	"github.com/cad/ovpm/pb"
	"github.com/urfave/cli"
)

var action string

func main() {
	app := cli.NewApp()
	app.Name = "ovpmd"
	app.Usage = "OpenVPN Manager Daemon"
	app.Version = ovpm.Version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:  "port",
			Usage: "port number for daemon to listen on",
		},
	}
	app.Before = func(c *cli.Context) error {
		logrus.SetLevel(logrus.InfoLevel)
		if c.GlobalBool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		ovpm.SetupDB()
		return nil
	}
	app.After = func(c *cli.Context) error {
		ovpm.CeaseDB()
		return nil
	}
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		if port == "" {
			port = "9090"
		}
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			logrus.Fatalf("could not listen to port %s: %v", port, err)
		}
		s := grpc.NewServer()
		pb.RegisterUserServiceServer(s, &api.UserService{})
		pb.RegisterVPNServiceServer(s, &api.VPNService{})
		logrus.Infof("OVPM is running :%s ...", port)
		s.Serve(lis)
		return nil
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
