package gounit

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
)

var (
	ErrGenerateHeader        = GenericError("failed to write header: %v")
	ErrGenerateTest          = GenericError("failed to write test: %v")
	ErrFuncNotFound          = GenericError("unable to find a function declaration")
	ErrFailedToParseInFile   = GenericError("failed to parse input file: %v")
	ErrFailedToParseOutFile  = GenericError("failed to parse output file: %v")
	ErrFailedToOpenInFile    = GenericError("failed to open input file: %v")
	ErrFailedToOpenOutFile   = GenericError("failed to open output file: %v")
	ErrFailedToCreateOutFile = GenericError("failed to create output file: %v")
	ErrInputFileDoesNotExist = GenericError("input file does not exist")
	ErrSeekFailed            = GenericError("failed to seek: %v")
	ErrFixImports            = GenericError("failed to fix imports: %v")
	ErrWriteTest             = GenericError("failed to write generated test: %v")
)

type LinesNumbers []int
type FunctionsList []string

type Options struct {
	Lines      LinesNumbers
	Functions  FunctionsList
	InputFile  string
	OutputFile string
	Comment    string
	Template   string
	All        bool
	UseJSON    bool
	UseStdin   bool
	UseStdout  bool
}

//GenerateCommand implements Command interface
type GenerateCommand struct {
	Options Options
	fs      *flag.FlagSet
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
		gc.fs.Var(&o.Lines, "l", "comma-separated line numbers (starting with 1) to look for the function declarations")
		gc.fs.Var(&o.Functions, "f", "comma-separated function names to generate tests for")
	}

	return gc.fs
}

func (gc *GenerateCommand) Run(args []string, stdout, stderr io.Writer) error {
	if err := gc.FlagSet().Parse(args); err != nil {
		return CommandLineError(err.Error())
	}

	options := gc.Options

	options.All = (len(options.Lines) == 0 && len(options.Functions) == 0)

	if options.InputFile == "" {
		return CommandLineError("missing input file")
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
			return ErrFailedToOpenInFile.Format(err)
		}

		if !options.UseStdin {
			return ErrInputFileDoesNotExist
		}

		r = os.Stdin
	}

	outFile, err := os.OpenFile(options.OutputFile, os.O_RDWR, 0600)
	if err != nil {
		if !os.IsNotExist(err) {
			ErrFailedToOpenOutFile.Format(err)
		}
	} else {
		w = outFile
		testSrc = outFile
	}

	options.Template, err = getDefaultTemplate()
	if err != nil {
		return err
	}

	generator, err := NewGenerator(options, r, testSrc)
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
			return ErrSeekFailed.Format(err)
		}
	}

	if options.UseStdout {
		w = os.Stdout
	}

	if b := buf.Bytes(); len(b) > 0 { //some code has been generated
		if w == nil {
			if w, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				ErrFailedToCreateOutFile.Format(err)
			}
			defer w.Close()
		}

		if _, err = w.Write(b); err != nil {
			ErrWriteTest.Format(err)
		}
	}

	return nil
}

func (gc *GenerateCommand) processJSON(r io.Reader, w io.Writer) error {
	var jo Request

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

		opt := Options{
			InputFile:  jo.InputFilePath,
			OutputFile: jo.OutputFilePath,
			Comment:    jo.Comment,
			Lines:      LinesNumbers(jo.Lines),
		}

		opt.Template, err = getDefaultTemplate()
		if err != nil {
			return err
		}

		generator, err := NewGenerator(opt, inputFile, outputFile)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer([]byte{})

		if err := generator.Write(b); err != nil {
			return err
		}

		if err := encoder.Encode(Response{GeneratedCode: b.String()}); err != nil {
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
