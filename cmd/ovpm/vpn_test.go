package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVPNCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	err := app.Run([]string{"ovpm", "vpn"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "status, s") {
		t.Fatal("subcommand missing 'status, s'")
	}

	if !strings.Contains(output.String(), "init, i") {
		t.Fatal("subcommand missing 'init, i'")
	}

	if !strings.Contains(output.String(), "update, u") {
		t.Fatal("subcommand missing 'update, u'")
	}

	if !strings.Contains(output.String(), "restart, r") {
		t.Fatal("subcommand missing 'restart, r'")
	}
}
