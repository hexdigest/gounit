package gounit

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"
)

type testCmd string

func (t testCmd) Description() string {
	return string(t)
}

func (t testCmd) Usage() string {
	return string(t)
}

func (t testCmd) Run(args []string, out, err io.Writer) error {
	return nil
}

func (t testCmd) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("testcmd", flag.ContinueOnError)
}

func TestRegisterCommand(t *testing.T) {
	RegisterCommand("gen", testCmd("gen"))

	_, ok := commands["gen"]
	if !ok {
		t.Fatalf("\"gen\" key is not present in the commands map")
	}
}

func TestGetCommand(t *testing.T) {
	commands["gen"] = testCmd("gen")
	if GetCommand("gen") == nil {
		t.Fatalf("expected non-nil value")
	}
}

func TestUsage(t *testing.T) {
	commands["_generate_"] = testCmd("gen")

	b := bytes.NewBuffer([]byte{})
	Usage(b)

	if !strings.Contains(b.String(), "_generate_") {
		t.Fatalf("no _generate_ command name found in the output")
	}
}
