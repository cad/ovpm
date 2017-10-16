package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUserCmd(t *testing.T) {
	app := NewApp()

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
