package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm/errors"
	"github.com/urfave/cli"

	"google.golang.org/grpc"
)

func emitToFile(filePath, content string, mode uint) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Cannot create file %s: %v", filePath, err)

	}
	if mode != 0 {
		file.Chmod(os.FileMode(mode))
	}
	defer file.Close()
	fmt.Fprintf(file, content)
	return nil
}

func getConn(port string) *grpc.ClientConn {
	if port == "" {
		port = "9090"
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%s", port), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("fail to dial: %v", err)
	}
	return conn
}

// grpcConnect receives a rpc server url and makes a connection to the
// GRPC server.
func grpcConnect(rpcServURL *url.URL) (*grpc.ClientConn, error) {
	// Ensure rpcServURL host part contains a localhost addr only.
	if !isLoopbackURL(rpcServURL) {
		return nil, errors.MustBeLoopbackURL(rpcServURL)
	}

	conn, err := grpc.Dial(rpcServURL.Host, grpc.WithInsecure())
	if err != nil {
		return nil, errors.UnknownSysError(err)
	}

	return conn, nil
}

// isLoopbackURL is a utility function that determines whether the
// given url.URL's host part resolves to a loopback ip addr or not.
func isLoopbackURL(u *url.URL) bool {
	// Resolve url to ip addresses.
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Ensure all resolved ip addrs are loopback addrs.
	for _, ip := range ips {
		if !ip.IsLoopback() {
			return false
		}
	}

	return true
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

// Prints the received message followed by the usage string.
func failureMsg(c *cli.Context, msg string) {
	fmt.Printf(msg)
	fmt.Println()
	fmt.Println(cli.ShowSubcommandHelp(c))
}
