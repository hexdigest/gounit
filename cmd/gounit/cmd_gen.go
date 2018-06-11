package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hexdigest/gounit"
)

type LinesNumbers []int
type FunctionsList []string

//GenerateCommand implements Command interface
type GenerateCommand struct {
	Options gounit.Options
	fs      *flag.FlagSet
	lines   LinesNumbers
	funcs   FunctionsList
}

//Description implements Command interface
func (gc *GenerateCommand) Description() string {
	return "generate test stub(s)"
}

func (gc *GenerateCommand) Usage() string {
	return "usage: gounit gen [-i input file] [-o output file] [-all | -l lines | -f functions]"
}

func (gc *GenerateCommand) FlagSet() *flag.FlagSet {
	o := &gc.Options

	if gc.fs == nil {
		gc.fs = &flag.FlagSet{}
		gc.fs.BoolVar(&o.All, "all", true, "generate tests for all functions")
		gc.fs.BoolVar(&o.UseJSON, "json", false, "read JSON-encoded input parameters from stdin\nplease see http://github.com/hexdigest/gounit for details")
		gc.fs.BoolVar(&o.UseStdin, "stdin", false, "use stdin rather than reading the input file")
		gc.fs.BoolVar(&o.UseStdout, "stdout", false, "use stdout rather than writing to the output file")
		gc.fs.StringVar(&o.InputFile, "i", "", "input file name")
		gc.fs.StringVar(&o.OutputFile, "o", "", "output file name (optional)")
		gc.fs.StringVar(&o.Comment, "c", "", "comment that will be inserted into the generated test")
		gc.fs.Var(&gc.lines, "l", "comma-separated line numbers (starting with 1) to look for the function declarations")
		gc.fs.Var(&gc.funcs, "f", "comma-separated function names to generate tests for")
	}

	return gc.fs
}

func (gc *GenerateCommand) Run(args []string, stdout, stderr io.Writer) error {
	if err := gc.FlagSet().Parse(args); err != nil {
		return gounit.CommandLineError(err.Error())
	}

	options := gc.Options
	options.Lines = []int(gc.lines)
	options.Functions = []string(gc.funcs)

	options.All = (len(options.Lines) == 0 && len(options.Functions) == 0)

	if options.InputFile == "" {
		return gounit.CommandLineError("missing input file")
	}

	if options.OutputFile == "" {
		if strings.HasSuffix(options.InputFile, ".go") {
			chunks := strings.Split(options.InputFile, ".")
			options.OutputFile = strings.Join(chunks[0:len(chunks)-1], ".")
		} else {
			options.OutputFile = options.InputFile
		}
		options.OutputFile += "_test.go"
	}

	if options.UseJSON {
		if err := gc.processJSON(os.Stdin, stdout); err != nil {
			return err
		}
	}

	var (
		r, testSrc io.Reader
		w          io.WriteCloser
		err        error
		buf        = bytes.NewBuffer([]byte{})
	)

	r, err = os.Open(options.InputFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return gounit.ErrFailedToOpenInFile.Format(err)
		}

		if !options.UseStdin {
			return gounit.ErrInputFileDoesNotExist
		}

		r = os.Stdin
	}

	outFile, err := os.OpenFile(options.OutputFile, os.O_RDWR, 0600)
	if err != nil {
		if !os.IsNotExist(err) {
			gounit.ErrFailedToOpenOutFile.Format(err)
		}
	} else {
		w = outFile
		testSrc = outFile
	}

	options.Template, err = getDefaultTemplate()
	if err != nil {
		return err
	}

	generator, err := gounit.NewGenerator(options, r, testSrc)
	if err != nil {
		return err
	}

	if err := generator.Write(buf); err != nil {
		return err
	}

	//rewind output file back to write from the beginning without
	//re-opening the file
	if seeker, ok := w.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return gounit.ErrSeekFailed.Format(err)
		}
	}

	if options.UseStdout {
		w = os.Stdout
	}

	if b := buf.Bytes(); len(b) > 0 { //some code has been generated
		if w == nil {
			if w, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				gounit.ErrFailedToCreateOutFile.Format(err)
			}
			defer w.Close()
		}

		if _, err = w.Write(b); err != nil {
			gounit.ErrWriteTest.Format(err)
		}
	}

	return nil
}

func (gc *GenerateCommand) processJSON(r io.Reader, w io.Writer) error {
	var jo gounit.Request

	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r)

	for {
		err := decoder.Decode(&jo)
		if err != nil {
			return err
		}

		inputFile := strings.NewReader(jo.InputFile)

		var outputFile io.Reader
		if len(jo.OutputFile) > 0 {
			outputFile = strings.NewReader(jo.OutputFile)
		}

		opt := gounit.Options{
			InputFile:  jo.InputFilePath,
			OutputFile: jo.OutputFilePath,
			Comment:    jo.Comment,
			Lines:      jo.Lines,
		}

		opt.Template, err = getDefaultTemplate()
		if err != nil {
			return err
		}

		generator, err := gounit.NewGenerator(opt, inputFile, outputFile)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer([]byte{})

		if err := generator.Write(b); err != nil {
			return err
		}

		if err := encoder.Encode(gounit.Response{GeneratedCode: b.String()}); err != nil {
			return err
		}
	}
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
