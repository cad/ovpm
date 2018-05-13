package main

import (
	"fmt"
	"net"

	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/errors"
	"github.com/urfave/cli"
)

// userListCmd lists existing VPN users.
//
// List includes additional information about users in addition to usernames
// such as; `IP`, `CREATED AT`, `VALID CRT`, `PUSH GW`.
var userListCmd = cli.Command{
	Name:    "list",
	Usage:   "List VPN users.",
	Aliases: []string{"l"},
	Action: func(c *cli.Context) error {
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return userListAction(fmt.Sprintf("grpc://localhost:%d", daemonPort))
	},
}

var userCreateCmd = cli.Command{
	Name:    "create",
	Usage:   "Create a VPN user.",
	Aliases: []string{"c"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "username, u",
			Usage: "username for the vpn user (required)",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "password for the vpn user (required)",
		},
		cli.BoolFlag{
			Name:  "no-gw",
			Usage: "don't push vpn server as default gateway for this user",
		},
		cli.StringFlag{
			Name:  "static",
			Usage: "ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "admin, a",
			Usage: "this user has admin rights",
		},
	},
	// userCreate action has two modes. Bulk mode
	Action: func(c *cli.Context) error {
		action = "user:create"

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and password.
		if username := c.String("username"); govalidator.IsNull(username) {
			return errors.EmptyValue("username", username)
		}
		if password := c.String("password"); govalidator.IsNull(password) {
			return errors.EmptyValue("password", password)
		}

		// Static IP addr holder.
		var ipAddr *net.IP

		// If static IP addr string is set by the user, then parse it as net.IP.
		if ipAddrStr := c.String("static"); !govalidator.IsNull(ipAddrStr) {
			// Validate the IP string.
			if !govalidator.IsIPv4(ipAddrStr) {
				err := errors.NotIPv4(ipAddrStr)
				exit(1)
				return err
			}
			// Parse and assign to the ipAddr variable.
			tmp := net.ParseIP(ipAddrStr)
			ipAddr = &tmp
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		// Call the action.
		return userCreateAction(
			fmt.Sprintf("grpc://localhost:%d", daemonPort),
			c.String("username"),
			c.String("password"),
			ipAddr,
			c.Bool("no-gw"),
			c.Bool("admin"),
		)
	},
}

// userUpdateCmd updates a user. It receives a set of flags.
// It has two modes; individual mode and bulk mode.
//
// Individual mode (default): In this mode update command acts on a single user
//   whose name is supplied via `username` flag.
//
// Bulk mode: In bulk mode update comand acts on all of the users. In order to
//   activate this mode one must supply asterisk (*) as the username. Then all
//   users will be updated after evaluating the rest of the received flags.
var userUpdateCmd = cli.Command{
	Name:    "update",
	Usage:   "Update a VPN user.",
	Aliases: []string{"u"},
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
		cli.StringFlag{
			Name:  "static",
			Usage: "ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "no-static",
			Usage: "do not set static ip address for the vpn user",
		},
		cli.BoolFlag{
			Name:  "admin",
			Usage: "this user has admin rights",
		},
		cli.BoolFlag{
			Name:  "no-admin",
			Usage: "this user has no admin rights",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:update"

		// inBulk means opeation needs to be done on
		// all users.
		var inBulk bool

		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and maybe password if set.
		if govalidator.IsNull(c.String("username")) {
			err := errors.EmptyValue("username", c.String("username"))
			exit(1)
			return err
		}

		// Check if bulk update is set.
		if c.String("username") == "*" {
			inBulk = true
		}

		// Set password if it's provided.
		var password *string
		if passwordStr := c.String("password"); len(passwordStr) > 0 {
			password = &passwordStr
		}

		// Set isStatic if it's provided.
		var isStatic *bool
		var ipAddr *net.IP

		// Check mutex options.
		if !govalidator.IsNull(c.String("static")) == true && inBulk {
			err := errors.ConflictingDemands("--static and --user * (bulk) options are mutually exclusive (can not be used together)")
			exit(1)
			return err
		}
		if !govalidator.IsNull(c.String("static")) == true && c.Bool("no-static") == true {
			err := errors.ConflictingDemands("--static and --no-static options are mutually exclusive (can not be used together)")
			exit(1)
			return err
		}
		if ipAddrStr := c.String("static"); !govalidator.IsNull(ipAddrStr) {
			// Validate the IP string.
			if !govalidator.IsIPv4(ipAddrStr) {
				err := errors.NotIPv4(ipAddrStr)
				exit(1)
				return err
			}
			// Set isStatic accordingly.
			isStaticVal := true
			isStatic = &isStaticVal

			// Parse and assign to the ipAddr variable.
			ipAddrVal := net.ParseIP(ipAddrStr)
			ipAddr = &ipAddrVal

		} else {
			if c.Bool("no-static") {
				tmp := false
				isStatic = &tmp
			}
		}

		// Set noGW if it's provided.
		var noGW *bool
		gwVal, noGWVal := c.Bool("gw"), c.Bool("no-gw")
		if gwVal == true && noGWVal == true {
			err := errors.ConflictingDemands("--gw and --no-gw options are mutually exclusive (can not be used together)")
			exit(1)
			return err
		}
		if gwVal {
			tmp := false
			noGW = &tmp
		}
		if noGWVal {
			tmp := true
			noGW = &tmp
		}

		// Set isAdmin if it's provided.
		var isAdmin *bool
		admin, noAdmin := c.Bool("admin"), c.Bool("no-admin")
		if admin == true && noAdmin == true {
			err := errors.ConflictingDemands("--admin and --no-admin options are mutually exclusive (can not be used together)")
			exit(1)
			return err
		}
		if admin {
			tmp := true
			isAdmin = &tmp
		}
		if noAdmin {
			tmp := false
			isAdmin = &tmp
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		// Call the action.
		return userUpdateAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("username"),
			password,
			ipAddr,
			isStatic,
			noGW,
			isAdmin,
			inBulk,
		)
	},
}

var userDeleteCmd = cli.Command{
	Name:    "delete",
	Usage:   "Delete a VPN user.",
	Aliases: []string{"d"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "username of the vpn user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:delete"
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and password.
		if username := c.String("user"); govalidator.IsNull(username) {
			return errors.EmptyValue("username", username)
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return userDeleteAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("user"))
	},
}

var userRenewCmd = cli.Command{
	Name:    "renew",
	Usage:   "Renew VPN user certificates.",
	Aliases: []string{"r"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "username of the vpn user",
		},
	},
	Action: func(c *cli.Context) error {
		action = "user:renew"
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username and password.
		if username := c.String("user"); govalidator.IsNull(username) {
			return errors.EmptyValue("username", username)
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return userRenewAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("user"))
	},
}

var userGenconfigCmd = cli.Command{
	Name:    "genconfig",
	Usage:   "Generate client config for the user. (.ovpn file)",
	Aliases: []string{"g"},
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
		// Use default port if no port is specified.
		daemonPort := ovpm.DefaultDaemonPort
		if port := c.GlobalInt("daemon-port"); port != 0 {
			daemonPort = port
		}

		// Validate username.
		if username := c.String("user"); govalidator.IsNull(username) {
			return errors.EmptyValue("username", username)
		}

		// Set outPath if it's provided.
		var outPath *string
		if outPathVal := c.String("out"); !govalidator.IsNull(outPathVal) {
			outPath = &outPathVal
		}

		// If dry run, then don't call the action, just preprocess.
		if c.GlobalBool("dry-run") {
			return nil
		}

		return userGenconfigAction(fmt.Sprintf("grpc://localhost:%d", daemonPort), c.String("user"), outPath)
	},
}

func init() {
	app.Commands = append(app.Commands,
		cli.Command{
			Name:    "user",
			Usage:   "User Operations",
			Aliases: []string{"u"},
			Subcommands: []cli.Command{
				userListCmd,
				userCreateCmd,
				userUpdateCmd,
				userDeleteCmd,
				userRenewCmd,
				userGenconfigCmd,
			},
		},
	)
}
