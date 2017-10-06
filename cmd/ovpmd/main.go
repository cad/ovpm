//go:generate go-bindata template/

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api"
	"github.com/urfave/cli"
)

var action string
var db *ovpm.DB

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
			Usage: "port number for gRPC API daemon",
		},
		cli.StringFlag{
			Name:  "web-port",
			Usage: "port number for the REST API daemon",
		},
	}
	app.Before = func(c *cli.Context) error {
		logrus.SetLevel(logrus.InfoLevel)
		if c.GlobalBool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		db = ovpm.CreateDB("sqlite3", "")
		return nil
	}
	app.After = func(c *cli.Context) error {
		db.Cease()
		return nil
	}
	app.Action = func(c *cli.Context) error {
		port := c.String("port")
		if port == "" {
			port = "9090"
		}

		webPort := c.String("web-port")
		if webPort == "" {
			webPort = "8080"
		}

		s := newServer(port, webPort)
		s.start()
		s.waitForInterrupt()
		s.stop()
		return nil
	}
	app.Run(os.Args)
}

type server struct {
	grpcPort   string
	lis        net.Listener
	grpcServer *grpc.Server
	restServer http.Handler
	restCancel context.CancelFunc
	restPort   string
	signal     chan os.Signal
	done       chan bool
}

func newServer(port, webPort string) *server {
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
		// NOTE(cad): gRPC endpoint listens on localhost. This is important
		// because we don't authanticate requests coming from localhost.
		// So gRPC endpoint should never listen on something else then
		// localhost.
		lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
		if err != nil {
			logrus.Fatalf("could not listen to port %s: %v", port, err)
		}

		rpcServer := api.NewRPCServer()
		restServer, restCancel, err := api.NewRESTServer(port)
		if err != nil {
			logrus.Fatalf("could not get new rest server :%v", err)
		}

		return &server{
			lis:        lis,
			grpcServer: rpcServer,
			restServer: restServer,
			restCancel: context.CancelFunc(restCancel),
			restPort:   webPort,
			signal:     sigs,
			done:       done,
			grpcPort:   port,
		}
	}
	return &server{}

}

func (s *server) start() {
	logrus.Infof("OVPM %s is running gRPC:%s, REST:%s ...", ovpm.Version, s.grpcPort, s.restPort)
	go s.grpcServer.Serve(s.lis)
	go http.ListenAndServe(":"+s.restPort, s.restServer)
	ovpm.StartVPNProc()
}

func (s *server) stop() {
	logrus.Info("OVPM is shutting down ...")
	s.grpcServer.Stop()
	s.restCancel()
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

func increasePort(p string) string {
	i, err := strconv.Atoi(p)
	if err != nil {
		logrus.Panicf(fmt.Sprintf("can't convert %s to int: %v", p, err))

	}
	i++
	return fmt.Sprintf("%d", i)
}
