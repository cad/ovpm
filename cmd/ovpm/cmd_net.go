package main

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/errors"
	"github.com/urfave/cli"
)

var netDefineCommand = cli.Command{
	Name:    "def",
	Aliases: []string{"d"},
	Usage:   "Define a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cidr, c",
			Usage: "CIDR of the network",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of the network",
		},
		cli.StringFlag{
			Name:  "type, t",
			Usage: "type of the network (see $ovpm net types)",
		},
		cli.StringFlag{
			Name:  "via, v",
			Usage: "if network type is route, via represents route's gateway",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:create"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate network name.
		if netName := c.String("name"); govalidator.IsNull(netName) {
			err := errors.EmptyValue("net", netName)
			exit(1)
			return err
		}

		// Validate network types.
		if netType := c.String("type"); !ovpm.IsNetworkType(netType) {
			err := errors.NotValidNetworkType("type", netType)
			exit(1)
			return err
		}

		// Validate if via can be set.
		if netVia := c.String("via"); !govalidator.IsNull(netVia) {
			if ovpm.NetworkTypeFromString(c.String("type")) != ovpm.ROUTE {
				err := errors.ConflictingDemands("--via flag can only be used with --type ROUTE")
				exit(1)
				return err
			}
		}

		// Validate network CIDR.
		if netCIDR := c.String("cidr"); !govalidator.IsCIDR(netCIDR) {
			err := errors.NotCIDR(netCIDR)
			exit(1)
			return err
		}

		var via *string
		if !govalidator.IsNull(c.String("via")) {
			tmp := c.String("via")
			via = &tmp
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return netDefAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("name"), c.String("cidr"), c.String("type"), via)
	},
}

var netListCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "List defined networks.",
	Action: func(c *cli.Context) error {
		action = "net:list"
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		return netListAction(fmt.Sprintf("grpc://localhost:%d", daemonPort))
	},
}

var netTypesCommand = cli.Command{
	Name:    "types",
	Aliases: []string{"t"},
	Usage:   "Show available network types.",
	Action: func(c *cli.Context) error {
		action = "net:types"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return netTypesAction(fmt.Sprintf("grpc://localhost:%d", daemonPort))
	},
}

var netUndefineCommand = cli.Command{
	Name:    "undef",
	Aliases: []string{"u"},
	Usage:   "Undefine an existing network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:delete"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate network name.
		if networkName := c.String("net"); govalidator.IsNull(networkName) {
			err := errors.EmptyValue("net", networkName)
			exit(1)
			return err
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return netUndefAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("net"))
	},
}

var netAssociateCommand = cli.Command{
	Name:    "assoc",
	Aliases: []string{"a"},
	Usage:   "Associate a user with a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},

		cli.StringFlag{
			Name:  "user, u",
			Usage: "name of the user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:associate"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and network name.
		if netName := c.String("net"); govalidator.IsNull(netName) {
			err := errors.EmptyValue("network", netName)
			exit(1)
			return err
		}
		if username := c.String("user"); govalidator.IsNull(username) {
			err := errors.EmptyValue("username", username)
			exit(1)
			return err
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return netAssocAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("net"), c.String("user"))
	},
}

var netDissociateCommand = cli.Command{
	Name:    "dissoc",
	Aliases: []string{"di"},
	Usage:   "Dissociate a user from a network.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net, n",
			Usage: "name of the network",
		},

		cli.StringFlag{
			Name:  "user, u",
			Usage: "name of the user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "net:dissociate"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and network name.
		if netName := c.String("net"); govalidator.IsNull(netName) {
			err := errors.EmptyValue("network", netName)
			exit(1)
			return err
		}
		if username := c.String("username"); govalidator.IsNull(username) {
			err := errors.EmptyValue("username", username)
			exit(1)
			return err
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return netDissocAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("net"), c.String("user"))
	},
}

func init() {
	app.Commands = append(app.Commands,
		cli.Command{
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
	)
}
