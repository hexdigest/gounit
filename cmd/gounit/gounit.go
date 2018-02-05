package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"

	"github.com/hexdigest/gounit"
	"golang.org/x/tools/imports"
)

const (
	//application exit codes
	ecInputErr           = 3 //something's wrong with the input
	ecOutputErr          = 4 //generation/write error
	ecTestIsAlreadyExist = 5
)

func main() {
	options := gounit.GetOptions(os.Args[1:], os.Stdout, os.Stderr, os.Exit)

	var (
		r      io.Reader
		w      io.Writer
		err    error
		append bool
		fs     = token.NewFileSet()
		buf    = bytes.NewBuffer([]byte{})
	)

	r, err = os.Open(options.InputFile)
	if err != nil {
		if !os.IsNotExist(err) {
			exit(ecInputErr, "failed to open input file: %v", err)
		}

		if !options.UseStdin {
			exit(ecInputErr, "input file does not exist")
		}

		r = os.Stdin
	}

	file, err := parser.ParseFile(fs, options.InputFile, r, 0)
	if err != nil {
		exit(ecInputErr, "failed to parse file: %v", err)
	}

	foundFunc, err := gounit.FindSourceFunc(fs, file, options)
	if err != nil {
		exit(ecInputErr, "%v", err)
	}

	outFile, err := os.OpenFile(options.OutputFile, os.O_RDWR, 0600)
	if err != nil {
		if !os.IsNotExist(err) {
			exit(ecOutputErr, "failed to open output file: %v", err)
		}

		if !options.UseStdout {
			if outFile, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				exit(ecOutputErr, "failed to create output file: %v", err)
			}
		}
	} else {
		//using TeeReader to read the contents of the output file into the write buffer
		//so we can just append a new test to the end of the buffer
		tr := io.TeeReader(outFile, buf)
		append = true

		isExist, err := gounit.IsTestExist(fs, tr, foundFunc, options)
		if err != nil {
			exit(ecOutputErr, "failed to check if test is already present: %v", err)
		}

		if isExist {
			exit(ecTestIsAlreadyExist, "test is already exist")
		}

		if _, err := outFile.Seek(0, 0); err != nil {
			exit(ecOutputErr, "failed to seek: %v", err)
		}
	}
	w = outFile
	defer outFile.Close()

	if options.UseStdout {
		w = os.Stdout
	}

	generator := gounit.NewGenerator(fs, foundFunc, options)

	if !append {
		if err := generator.WriteHeader(buf, file.Name.String(), file.Imports); err != nil {
			exit(ecOutputErr, "failed to write header: %v", err)
		}
	}

	if err := generator.WriteTest(buf); err != nil {
		exit(ecOutputErr, "failed to generate code: %v", err)
	}

	formattedSource, err := imports.Process(options.OutputFile, buf.Bytes(), nil)
	if err != nil {
		exit(ecOutputErr, "failed to fix imports %s\n: %v", string(buf.Bytes()), err)
	}

	if _, err = w.Write(formattedSource); err != nil {
		exit(ecOutputErr, "failed to write generated test: %v", err)
	}
}

func exit(exitCode int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(exitCode)
}
