package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hexdigest/gounit"
)

var eof = []byte("<cli_example_eof>")

func main() {
	var inputFileName, outputFileName, lines, comment string

	cmd := exec.Command("gounit", "-cli")
	guStdout, err := cmd.StdoutPipe()
	if err != nil {
		die("failed to open stdout pipe: %v", err)
	}

	guStdin, err := cmd.StdinPipe()
	if err != nil {
		die("failed to open stdin pipe: %v", err)
	}

	guStderr, err := cmd.StderrPipe()
	if err != nil {
		die("failed to open stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		die("failed to start gounit: %v", err)
	}

	b := bufio.NewReader(guStdout)

	go func() {
		message, err := ioutil.ReadAll(guStderr)
		if err != nil {
			die("failed to read from gounit stderr stream: %v", err)
		}
		err = cmd.Wait()
		die("gounit exited with error: %v, %s", err, string(message))
	}()

	cli := gounit.NewCLI(os.Stdin, os.Stdout)

	for {
		cli.Read("input file name", &inputFileName)
		cli.Read("output file name (file may not exist)", &outputFileName)
		cli.Read("line numbers", &lines)
		cli.Read("comment", &comment)

		inBytes, err := read(inputFileName)
		if err != nil {
			die("failed to read input file: %v", err)
		}

		var testsBytes []byte
		testsBytes, err = read(outputFileName)
		if err != nil && !os.IsNotExist(err) {
			die("failed to read output file: %v", err)
		}

		b.ReadString('\n') //skipping gounit prompt
		fmt.Fprintln(guStdin, inputFileName)

		b.ReadString('\n')
		fmt.Fprintln(guStdin, outputFileName)

		b.ReadString('\n')
		fmt.Fprintln(guStdin, lines)

		b.ReadString('\n')
		fmt.Fprintln(guStdin, comment)

		b.ReadString('\n')
		fmt.Fprintln(guStdin, string(eof))

		b.ReadString('\n')
		guStdin.Write(inBytes)
		guStdin.Write(eof)

		b.ReadString('\n')
		guStdin.Write(testsBytes)
		guStdin.Write(eof)

		generatedCode, err := readUntilEOF(guStdout)
		if err != nil {
			die("failed to read generated file: %v", err)
		}

		showOutput(testsBytes, generatedCode, comment)
	}
}

func showOutput(testsBefore, testsAfter []byte, comment string) {
	linesBefore := getLinesCount(testsBefore)
	scanner := bufio.NewScanner(bytes.NewReader(testsAfter))
	ln := 1
	found := false
	for scanner.Scan() {
		format := "%5d %s\n"
		line := scanner.Text()
		if ln > linesBefore && strings.Contains(line, comment) && !found {
			format = ">%4d %s\n"
			found = true
		}
		fmt.Printf(format, ln, line)
		ln++
	}
}

func getLinesCount(b []byte) int {
	l := 0
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		l++
	}

	return l
}

func read(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func readUntilEOF(r io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(
		func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.Index(data, eof); i >= 0 {
				return i + len(eof), data[0:i], nil
			}
			if atEOF {
				return len(data), data, nil
			}
			return 0, nil, nil
		},
	)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan failed: %v", err)
	}

	return scanner.Bytes(), nil
}

func die(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
