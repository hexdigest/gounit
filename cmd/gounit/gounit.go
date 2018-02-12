package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hexdigest/gounit"
)

type exitFunc func(int)

func main() {
	options := gounit.GetOptions(os.Args[1:], os.Stdout, os.Stderr, os.Exit)
	if options.UseCLI {
		if err := interactive(os.Stdin, os.Stdout); err != nil {
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

func interactive(r io.Reader, w io.Writer) error {
	cli := gounit.NewCLI(r, w)
	for {
		var (
			opt   gounit.Options
			lines string
			eof   string
		)

		if err := cli.Read("input file name", &opt.InputFile); err != nil {
			return err
		}
		if err := cli.Read("output file name", &opt.OutputFile); err != nil {
			return err
		}
		if err := cli.Read("lines numbers with functions declarations", &lines); err != nil {
			return err
		}
		if err := opt.Lines.Set(lines); err != nil {
			return err
		}
		if err := cli.Read("TODO comment", &opt.Comment); err != nil {
			return err
		}

		if err := cli.Read("EOF sequence", &eof); err != nil {
			return err
		}

		inBytes, err := cli.ReadUntil("input file contents", eof)
		if err != nil {
			return gounit.ErrFailedToOpenInFile.Format(err)
		}

		outBytes, err := cli.ReadUntil("output file contents", eof)
		if err != nil {
			return gounit.ErrFailedToOpenOutFile.Format(err)
		}

		inputFile := bytes.NewBuffer(inBytes)

		var outputFile io.Reader
		if len(outBytes) != 0 {
			outputFile = bytes.NewBuffer(outBytes)
		}

		generator, err := gounit.NewGenerator(opt, inputFile, outputFile)
		if err != nil {
			exit(err)
		}

		generator.Write(w)
		if _, err := w.Write([]byte(eof)); err != nil {
			exit(gounit.ErrWriteTest.Format(err))
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
