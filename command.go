package gounit

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

type Command interface {
	//FlagSet returns command specific flag set. If command doesn't have any flags nil should be returned.
	FlagSet() *flag.FlagSet

	//Run runs command
	Run(args []string, stdout, stderr io.Writer) error

	//Description returns short description of a command that is shown in the help message
	Description() string

	//Usage line
	Usage() string
}

var commands = map[string]Command{}

func RegisterCommand(name string, cmd Command) {
	commands[name] = cmd
	if fs := cmd.FlagSet(); fs != nil {
		fs.Init("", flag.ContinueOnError)
		fs.SetOutput(ioutil.Discard)
	}
}

func GetCommand(name string) Command {
	return commands[name]
}

func Usage(w io.Writer) {
	names := []string{}
	for name := range commands {
		names = append(names, name)
	}

	sort.Strings(names)

	fmt.Fprintf(w, `GoUnit is a tool for generating test stubs for testing Go source code

Usage:

	gounit command [arguments]

The commands are:

`)

	for _, name := range names {
		fmt.Fprintf(w, "\t%-10s\t%s\n", name, commands[name].Description())
	}

	fmt.Fprintf(w, `
Use "gounit help [command]" for more information about a command.
`)
}
