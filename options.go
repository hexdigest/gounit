package gounit

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type LinesNumbers []int
type FunctionsList []string

//Options contains parsed command line options
type Options struct {
	Lines      LinesNumbers
	Functions  FunctionsList
	InputFile  string
	OutputFile string
	Comment    string
	UseCLI     bool
	UseStdin   bool
	UseStdout  bool
}

//Set implements flag.Value interface
func (ln *LinesNumbers) Set(value string) error {
	chunks := strings.Split(value, ",")
	for _, chunk := range chunks {
		l, err := strconv.ParseUint(chunk, 10, 64)
		if err != nil {
			return fmt.Errorf("expected unsigned int, got: %s", chunk)
		}

		*ln = append(*ln, int(l))
	}

	return nil
}

//String implements flag.Value interface
func (ln *LinesNumbers) String() string {
	return fmt.Sprintf("%d", []int(*ln))
}

//Include checks if the given line is in the LinesNumbers slice
func (ln LinesNumbers) Include(line int) bool {
	for _, l := range ln {
		if l == line {
			return true
		}
	}
	return false
}

var regexpIdent = regexp.MustCompile("^([a-zA-Z_][a-zA-Z0-9]*|\\*)$")

//Set implements flag.Value interface
func (fl *FunctionsList) Set(value string) error {
	chunks := strings.Split(value, ",")
	for _, chunk := range chunks {
		function := strings.TrimSpace(chunk)
		if !regexpIdent.MatchString(function) {
			return fmt.Errorf("bad function name: %s", function)
		}

		*fl = append(*fl, function)
	}

	return nil
}

//String implements flag.Value interface
func (fl *FunctionsList) String() string {
	return fmt.Sprintf("%s", []string(*fl))
}

//Include checks if the given line is in the LinesNumbers slice
func (fl FunctionsList) Include(function string) bool {
	for _, f := range fl {
		if f == function {
			return true
		}
	}
	return false
}

type exitFunc func(int)

//GetOptions parses arguments and returns Options struct on success, otherwise
//writes error message to the stderr writer and calls exit function
func GetOptions(arguments []string, stdout, stderr io.Writer, exit exitFunc) Options {
	var (
		lines      LinesNumbers
		functions  FunctionsList
		flagset    = flag.NewFlagSet("gounit", flag.ExitOnError)
		showHelp   = flagset.Bool("h", false, "display this help text and exit")
		cli        = flagset.Bool("cli", false, "interactive mode")
		useStdin   = flagset.Bool("stdin", false, "use stdin rather than reading the input file")
		useStdout  = flagset.Bool("stdout", false, "use stdout rather than writing to the output file")
		inputFile  = flagset.String("i", "", "input file")
		outputFile = flagset.String("o", "", "output file (optional)")
		comment    = flagset.String("c", "", "comment that will be inserted into the generated test")
	)
	flagset.Var(&lines, "l", "comma-separated line numbers (starting with 1) to look for the functions' declarations")
	flagset.Var(&functions, "f", "comma-separated functions' names to generate tests for")
	flagset.Parse(arguments)

	if *showHelp {
		flagset.SetOutput(stdout)
		flagset.Usage()
		exit(ExitCodeOK)
	}

	if *cli {
		return Options{UseCLI: true}
	}

	var errors []string
	if len(lines) == 0 && len(functions) == 0 {
		errors = append(errors, "missing line numbers or function names")
	}

	if *inputFile == "" {
		errors = append(errors, "missing input file")
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
		exit(ExitCodeErrCommandLine)
	}

	return Options{
		Lines:      lines,
		Functions:  functions,
		InputFile:  *inputFile,
		OutputFile: *outputFile,
		Comment:    *comment,
		UseStdin:   *useStdin,
		UseStdout:  *useStdout,
	}
}
