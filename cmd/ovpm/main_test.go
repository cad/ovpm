package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cad/ovpm"
)

func TestMainCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	err := app.Run([]string{"ovpm"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), ovpm.Version) {
		t.Fatal("version is missing")
	}

	if !strings.Contains(output.String(), "user, u") {
		t.Fatal("subcommand missing 'user'")
	}

	if !strings.Contains(output.String(), "vpn, v") {
		t.Fatal("subcommand missing 'vpn'")
	}

	if !strings.Contains(output.String(), "net, n") {
		t.Fatal("subcommand missing 'net'")
	}

	if !strings.Contains(output.String(), "help, h") {
		t.Fatal("subcommand missing 'help'")
	}

	if !strings.Contains(output.String(), "--daemon-port") {
		t.Fatal("flag missing '--daemon-port'")
	}

	if !strings.Contains(output.String(), "--verbose") {
		t.Fatal("flag missing '--verbose'")
	}

	if !strings.Contains(output.String(), "--version") {
		t.Fatal("flag missing '--version'")
	}

}
