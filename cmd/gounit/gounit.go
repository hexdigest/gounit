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
			exit(gounit.ErrFailedToOpenInFile, err)
		}

		if !options.UseStdin {
			exit(gounit.ErrInputFileDoesNotExist)
		}

		r = os.Stdin
	}

	file, err := parser.ParseFile(fs, options.InputFile, r, 0)
	if err != nil {
		exit(gounit.ErrFailedToParseInFile, err)
	}

	funcDecl, err := gounit.FindSourceFunc(fs, file, options)
	if err != nil {
		exit(gounit.ErrFailedToFindSourceFunc, err)
	}

	if funcDecl == nil {
		exit(gounit.ErrFuncNotFound)
	}

	foundFunc := gounit.NewFunc(fs, funcDecl)

	outFile, err := os.OpenFile(options.OutputFile, os.O_RDWR, 0600)
	if err != nil {
		if !os.IsNotExist(err) {
			exit(gounit.ErrFailedToOpenOutFile, err)
		}

		if !options.UseStdout {
			if outFile, err = os.OpenFile(options.OutputFile, os.O_CREATE|os.O_WRONLY, 0600); err != nil {
				exit(gounit.ErrFailedToCreateOutFile, err)
			}
		}
	} else {
		//using TeeReader to read the contents of the output file into the write buffer
		//so we can just append a new test to the end of the buffer
		tr := io.TeeReader(outFile, buf)
		append = true

		isExist, err := gounit.IsTestExist(fs, tr, foundFunc, options)
		if err != nil {
			exit(gounit.ErrFailedToParseOutFile, err)
		}

		if isExist {
			exit(gounit.ErrTestIsAlreadyExist)
		}

		if _, err := outFile.Seek(0, 0); err != nil {
			exit(gounit.ErrSeekFailed, err)
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
			exit(gounit.ErrGenerateHeader, err)
		}
	}

	if err := generator.WriteTest(buf); err != nil {
		exit(gounit.ErrGenerateTest, err)
	}

	formattedSource, err := imports.Process(options.OutputFile, buf.Bytes(), nil)
	if err != nil {
		fmt.Println(string(buf.Bytes()))
		exit(gounit.ErrFixImports, err)
	}

	if _, err = w.Write(formattedSource); err != nil {
		exit(gounit.ErrWriteTest, err)
	}
}

func exit(e gounit.Error, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%v\n", e.Format(args...))
	os.Exit(e.Code())
}
