package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hexdigest/gounit"
)

func main() {
	options := gounit.GetOptions(os.Args[1:], os.Stdout, os.Stderr, os.Exit)
	if options.UseJSON {
		if err := processJSON(os.Stdin, os.Stdout); err != nil {
			exit(err)
		}
		return
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
			exit(gounit.ErrFailedToOpenInFile.Format(err))
		}

		if !options.UseStdin {
			exit(gounit.ErrInputFileDoesNotExist)
		}

		r = os.Stdin
	}

	outFile, err := os.OpenFile(options.OutputFile, os.O_RDWR, 0600)
	if err != nil {
		if !os.IsNotExist(err) {
			exit(gounit.ErrFailedToOpenOutFile.Format(err))
		}
	} else {
		w = outFile
		testSrc = outFile
	}

	generator, err := gounit.NewGenerator(options, r, testSrc)
	if err != nil {
		exit(err)
	}

	if err := generator.Write(buf); err != nil {
		exit(err)
	}

	//rewind output file back to write from the beginning without
	//re-opening the file
	if seeker, ok := w.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			exit(gounit.ErrSeekFailed.Format(err))
		}
	}

	if options.UseStdout {
		w = os.Stdout
	}

	if b := buf.Bytes(); len(b) > 0 { //some code has been generated
		if w == nil {
			if w, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				exit(gounit.ErrFailedToCreateOutFile.Format(err))
			}
			defer w.Close()
		}

		if _, err = w.Write(b); err != nil {
			exit(gounit.ErrWriteTest.Format(err))
		}
	}
}

func processJSON(r io.Reader, w io.Writer) error {
	var jo gounit.Request

	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r)

	for {
		if err := decoder.Decode(&jo); err != nil {
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
			Lines:      gounit.LinesNumbers(jo.Lines),
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

func exit(err error) {
	var code = gounit.ExitCodeErrGeneric
	if gerr, ok := err.(*gounit.Error); ok {
		code = gerr.Code()
	}

	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(code)
}
