package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hexdigest/gounit"
)

func init() {
	gounit.RegisterCommand("gen", &GenerateCommand{})
	gounit.RegisterCommand("template", &TemplateCommand{})
}

func main() {
	if len(os.Args) < 2 {
		gounit.Usage(os.Stderr)
		os.Exit(2)
	}

	flag.Parse()
	args := flag.Args()

	if args[0] == "help" {
		exitCode := help(args[1:], os.Stdout, os.Stderr)
		os.Exit(exitCode)
	}

	command := gounit.GetCommand(args[0])
	if command == nil {
		fmt.Fprintf(os.Stderr, "gounit: unknown subcommand %q\nRun 'gounit help' for usage.\n", args[0])
		os.Exit(2)
	}

	if err := command.Run(args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)

		if _, ok := err.(gounit.CommandLineError); ok {
			fmt.Fprintf(os.Stderr, "Run 'gounit help %s' for usage.\n", args[0])
			os.Exit(2)
		} else {
			os.Exit(1)
		}
	}
}

func help(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintf(stderr, "usage: gounit help command\n\nToo many arguments given.\n")
		return 2
	}

	if len(args) == 0 {
		gounit.Usage(stderr)
		return 0
	}

	command := gounit.GetCommand(args[0])
	if command == nil {
		fmt.Fprintf(stderr, "gounit: unknown subcommand %q\nRun 'gounit help' for usage.\n", args[0])
		return 2
	}

	fmt.Fprintf(stdout, "%s\n", command.Usage())

	if fs := command.FlagSet(); fs != nil {
		fs.SetOutput(stdout)
		fs.PrintDefaults()
	}

	return 0
}
