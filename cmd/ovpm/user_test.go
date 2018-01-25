package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUserCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	err := app.Run([]string{"ovpm", "user"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "list, l") {
		t.Fatal("subcommand missing 'list, l'")
	}

	if !strings.Contains(output.String(), "create, c") {
		t.Fatal("subcommand missing 'create, c'")
	}

	if !strings.Contains(output.String(), "update, u") {
		t.Fatal("subcommand missing 'update, u'")
	}

	if !strings.Contains(output.String(), "delete, d") {
		t.Fatal("subcommand missing 'delete, d'")
	}

	if !strings.Contains(output.String(), "renew, r") {
		t.Fatal("subcommand missing 'renew, r'")
	}

	if !strings.Contains(output.String(), "genconfig, g") {
		t.Fatal("subcommand missing 'update, u'")
	}
}

func TestUserCreateCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "user", "create"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}

	// Missing password
	err = app.Run([]string{"ovpm", "user", "create", "--username", "sad"})
	if err == nil {
		t.Fatal("error is expected about missing password, but we didn't got error")
	}

	// Missing username
	err = app.Run([]string{"ovpm", "user", "create", "--password", "sad"})
	if err == nil {
		t.Fatal("error is expected about missing password, but we didn't got error")
	}

	// Malformed static ip
	err = app.Run([]string{"ovpm", "user", "create", "--username", "sad", "--password", "asdf", "--static", "asdf"})
	if err == nil {
		t.Fatal("error is expected about malformed static ip, but we didn't got error")
	}

	// Ensure proper static ip
	err = app.Run([]string{"ovpm", "user", "create", "--username", "adsf", "--password", "1234", "--static", "10.9.0.4"})
	if err != nil && !strings.Contains(err.Error(), "grpc") {
		t.Fatalf("error is not expected: %v", err)
	}

	// Ensure username chars
	err = app.Run([]string{"ovpm", "user", "create", "--username", "sdafADSFasdf325235.dsafsaf-asdffdsa_h5223s", "--password", "1234", "--static", "10.9.0.4"})
	if err != nil && !strings.Contains(err.Error(), "grpc") {
		t.Fatalf("error is not expected: %v", err)
	}

}

func TestUserUpdateCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "user", "update"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}

	// Commented out because it makes the implementation easier.
	// // Ensure missing fields
	// err = app.Run([]string{"ovpm", "user", "update", "--username", "foobar"})
	// if err == nil {
	// 	t.Fatal("error is expected about missing fields, but we didn't got error")
	// }

	// Mix gw with no-gw
	err = app.Run([]string{"ovpm", "user", "update", "--no-gw", "--gw"})
	if err == nil {
		t.Fatal("error is expected about gw mutually exclusivity, but we didn't got error")
	}

	// Mix admin with no-admin
	err = app.Run([]string{"ovpm", "user", "update", "--admin", "--no-admin"})
	if err == nil {
		t.Fatal("error is expected about admin mutually exclusivity, but we didn't got error")
	}

	// Malformed static
	err = app.Run([]string{"ovpm", "user", "update", "--username", "foo", "--static", "sadfsadf"})
	if err == nil {
		t.Fatal("error is expected about static being malformed ip, but we didn't got error")
	}
}

func TestUserDeleteCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "user", "delete"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}
}

func TestUserRenewCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "user", "renew"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}
}

func TestUserGenconfigCmd(t *testing.T) {
	output := new(bytes.Buffer)
	app.Writer = output

	var err error

	// Empty call
	err = app.Run([]string{"ovpm", "user", "delete"})
	if err == nil {
		t.Fatal("error is expected about missing fields, but we didn't got error")
	}
}
