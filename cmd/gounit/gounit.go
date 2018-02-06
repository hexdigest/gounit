package main

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"os"

	"github.com/hexdigest/gounit"
	"golang.org/x/tools/imports"
)

func main() {
	options := gounit.GetOptions(os.Args[1:], os.Stdout, os.Stderr, os.Exit)

	var (
		r, testSrc io.Reader
		w          io.Writer
		err        error
		append     bool
		fs         = token.NewFileSet()
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

		if !options.UseStdout {
			if outFile, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				exit(gounit.ErrFailedToCreateOutFile.Format(err))
			}
		}
	} else {
		testSrc = io.TeeReader(outFile, buf)
		append = true
	}

	defer outFile.Close()

	w = outFile
	if options.UseStdout {
		w = os.Stdout
	}

	generator, err := gounit.NewGenerator(fs, options, r, testSrc)
	if err != nil {
		exit(err)
	}

	//rewind output file back
	if seeker, ok := w.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			exit(gounit.ErrSeekFailed.Format(err))
		}
	}

	if !append {
		if err := generator.WriteHeader(buf); err != nil {
			exit(gounit.ErrGenerateHeader.Format(err))
		}
	}

	if err := generator.WriteTests(buf); err != nil {
		exit(gounit.ErrGenerateTest.Format(err))
	}

	formattedSource, err := imports.Process(options.OutputFile, buf.Bytes(), nil)
	if err != nil {
		exit(gounit.ErrFixImports.Format(err))
	}

	if _, err = w.Write(formattedSource); err != nil {
		exit(gounit.ErrWriteTest.Format(err))
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
