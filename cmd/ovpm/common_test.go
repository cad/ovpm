package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func SetupTest() {
}

func init() {
	prevBeforeFunc := app.Before

	// Override dry-run flag ensuring it's set to true when testing.
	app.Before = func(c *cli.Context) error {
		if err := c.GlobalSet("dry-run", "true"); err != nil {
			fmt.Printf("can not set global flag 'dry-run' to true: %v\n", err)
			return err
		}

		return prevBeforeFunc(c)
	}
}
