package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hexdigest/gounit"
)

func main() {
	cmd := exec.Command("gounit", "-json")
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

	defer func() {
		message, err := ioutil.ReadAll(guStderr)
		if err != nil {
			die("failed to read from gounit stderr stream: %v", err)
		}
		err = cmd.Wait()
		die("gounit exited with error: %v, %s", err, string(message))
	}()

	var (
		decoder = json.NewDecoder(guStdout)
		encoder = json.NewEncoder(guStdin)
		s       string
	)
	for {
		var (
			request  gounit.Request
			response gounit.Response
			lines    gounit.LinesNumbers
		)
		read("input file path", &request.InputFilePath)
		if request.InputFile, err = readFile(request.InputFilePath); err != nil {
			die("failed to read input file: %v", err)
		}

		read("output file path(file may not exist)", &request.OutputFilePath)
		if request.OutputFile, err = readFile(request.OutputFilePath); err != nil && !os.IsNotExist(err) {
			die("failed to read output file: %v", err)
		}

		read("comment", &request.Comment)

		read("line numbers", &s)
		if err := lines.Set(s); err != nil {
			die("invalid lines numbers: %v", err)
		}

		request.Lines = []int(lines)

		if err := encoder.Encode(request); err != nil {
			die("failed to encode request: %v", err)
		}

		if err := decoder.Decode(&response); err != nil {
			if err != io.EOF {
				die("failed to decode response: %v", err)
			}
			return
		}

		if len(response.GeneratedCode) == 0 {
			fmt.Println("All tests for requested functions already exist")
		} else {
			showOutput(request.OutputFile, response.GeneratedCode, request.Comment)
		}
	}
}

func showOutput(testsBefore, testsAfter string, comment string) {
	linesBefore := getLinesCount(testsBefore)
	scanner := bufio.NewScanner(strings.NewReader(testsAfter))
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

func getLinesCount(text string) int {
	l := 0
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		l++
	}

	return l
}

func readFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func read(prompt string, dest interface{}) {
	fmt.Printf("%s:\n", prompt)
	if _, err := fmt.Scanln(dest); err != nil {
		die("failed to read %T: %v", dest, err)
	}
}

func die(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
