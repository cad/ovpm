//go:generate go-bindata template/

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		ovpm.SetupDB("sqlite3", "")
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
		s := newServer(port)
		s.start()
		s.waitForInterrupt()
		s.stop()
		return nil
	}
	app.Run(os.Args)
}

type server struct {
	port       string
	lis        net.Listener
	grpcServer *grpc.Server
	signal     chan os.Signal
	done       chan bool
}

func newServer(port string) *server {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	if !ovpm.Testing {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			logrus.Fatalf("could not listen to port %s: %v", port, err)
		}
		s := grpc.NewServer()
		pb.RegisterUserServiceServer(s, &api.UserService{})
		pb.RegisterVPNServiceServer(s, &api.VPNService{})
		pb.RegisterNetworkServiceServer(s, &api.NetworkService{})
		return &server{lis: lis, grpcServer: s, signal: sigs, done: done, port: port}
	}
	return &server{}

}

func (s *server) start() {
	logrus.Infof("OVPM is running :%s ...", s.port)
	go s.grpcServer.Serve(s.lis)
	ovpm.StartVPNProc()
}

func (s *server) stop() {
	logrus.Info("OVPM is shutting down ...")
	s.grpcServer.Stop()
	ovpm.StopVPNProc()
}

func (s *server) waitForInterrupt() {
	<-s.done
	go timeout(8 * time.Second)
}

func timeout(interval time.Duration) {
	time.Sleep(interval)
	log.Println("Timeout! Killing the main thread...")
	os.Exit(-1)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
