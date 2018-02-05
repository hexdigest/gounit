package gounit

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

//Options contains parsed command line options
type Options struct {
	LineNumber int
	InputFile  string
	OutputFile string
	Comment    string
	UseStdin   bool
	UseStdout  bool
}

type exitFunc func(int)

//GetOptions parses arguments and returns Options struct on success, otherwise
//writes error message to the stderr writer and calls exit function
func GetOptions(arguments []string, stdout, stderr io.Writer, exit exitFunc) Options {
	var (
		flagset    = flag.NewFlagSet("gounit", flag.ExitOnError)
		showHelp   = flagset.Bool("h", false, "display this help text and exit")
		useStdin   = flagset.Bool("stdin", false, "use stdin rather than reading the input file")
		useStdout  = flagset.Bool("stdout", false, "use stdout rather than writing to the output file")
		lineNumber = flagset.Uint("l", 0, "number of the line (starting with 1) with the function declaration")
		inputFile  = flagset.String("i", "", "input file (optional)")
		outputFile = flagset.String("o", "", "output file")
		comment    = flagset.String("c", "", "comment that will be inserted to the generated test")
	)

	flagset.Parse(arguments)

	if *showHelp {
		flagset.SetOutput(stdout)
		flagset.Usage()
		exit(0)
	}

	var errors []string
	if *lineNumber == 0 {
		errors = append(errors, "missing line number: -l")
	}

	if *inputFile == "" {
		errors = append(errors, "missing input file: -i")
	}

	if *outputFile == "" {
		if strings.HasSuffix(*inputFile, ".go") {
			chunks := strings.Split(*inputFile, ".")
			*outputFile = strings.Join(chunks[0:len(chunks)-1], ".")
		} else {
			*outputFile = *inputFile
		}
		*outputFile += "_test.go"
	}

	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(stderr, "%s\n", e)
		}
		flagset.Usage()
		exit(2)
	}

	return Options{
		LineNumber: int(*lineNumber),
		InputFile:  *inputFile,
		OutputFile: *outputFile,
		Comment:    *comment,
		UseStdin:   *useStdin,
		UseStdout:  *useStdout,
	}
}
