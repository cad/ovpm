package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNetCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	err := app.Run([]string{"ovpm", "net"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "list, l") {
		t.Fatal("subcommand missing 'list, l'")
	}

	if !strings.Contains(output.String(), "types, t") {
		t.Fatal("subcommand missing 'types, t'")
	}

	if !strings.Contains(output.String(), "def, d") {
		t.Fatal("subcommand missing 'undef, u'")
	}

	if !strings.Contains(output.String(), "assoc, a") {
		t.Fatal("subcommand missing 'assoc, a'")
	}

	if !strings.Contains(output.String(), "dissoc, di") {
		t.Fatal("subcommand missing 'dissoc, di'")
	}
}

func TestNetDefineCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "net", "def"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}

	// Missing type
	err = app.Run([]string{"ovpm", "net", "def", "--cidr", "192.168.1.1/24"})
	if err == nil {
		t.Fatal("error is expected about missing network type, but we didn't got error")
	}
	// Missing name
	err = app.Run([]string{"ovpm", "net", "def", "--type", "SERVERNET", "--cidr", "192.168.1.1/24"})
	if err == nil {
		t.Fatal("error is expected about missing network name, but we didn't got error")
	}

	// Incorrect type
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd", "--type", "SERVERNUT", "--cidr", "192.168.1.1/24"})
	if err == nil {
		t.Fatal("error is expected about incorrect server type, but we didn't got error")
	}

	// Incorrect use of via
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd", "--type", "SERVERNET", "--cidr", "192.168.1.1/24", "--via", "8.8.8.8"})
	if err == nil {
		t.Fatal("error is expected about incorrect use of via, but we didn't got error")
	}

	// Incorrect cidr format
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd", "--type", "SERVERNET", "--cidr", "192.168.1.1"})
	if err == nil {
		t.Fatal("error is expected about incorrect cidr format, but we didn't got error")
	}

	// Ensure ROUTE type use without --via
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd", "--type", "ROUTE", "--cidr", "192.168.1.1/24"})
	if err != nil && !strings.Contains(err.Error(), "grpc") {
		t.Fatalf("error is not expected: %v", err)
	}

	// Incorrect use of via
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd", "--type", "SERVERNET", "--cidr", "192.168.1.1/24", "--via", "8.8.8.8/24"})
	if err == nil {
		t.Fatal("error is expected about incorrect via format, but we didn't got error")
	}

	// Ensure network name alphanumeric and dot, underscore chars are allowed
	err = app.Run([]string{"ovpm", "net", "def", "--name", "asd.asdd5sa_fasA32", "--type", "ROUTE", "--cidr", "192.168.1.1/24"})
	if err != nil && !strings.Contains(err.Error(), "grpc") {
		t.Fatalf("error is not expected: %v", err)
	}

}

func TestNetUnDefineCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "net", "undef"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}

}

func TestAssocCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "net", "assoc"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}

	// Missing network name
	err = app.Run([]string{"ovpm", "net", "def", "--user", "asd"})
	if err == nil {
		t.Fatal("error is expected about missing network name, but we didn't got error")
	}

	// Missing username
	err = app.Run([]string{"ovpm", "net", "def", "--network", "asddsa"})
	if err == nil {
		t.Fatal("error is expected about missing username, but we didn't got error")
	}
}
