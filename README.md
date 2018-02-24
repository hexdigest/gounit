# GoUnit [![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](https://github.com/hexdigest/gounit/blob/master/LICENSE) [![Build Status](https://travis-ci.org/hexdigest/gounit.svg?branch=master)](https://travis-ci.org/hexdigest/gounit) [![Coverage Status](https://coveralls.io/repos/github/hexdigest/gounit/badge.svg?branch=master)](https://coveralls.io/github/hexdigest/gounit?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hexdigest/gounit)](https://goreportcard.com/report/github.com/hexdigest/gounit) [![GoDoc](https://godoc.org/github.com/hexdigest/gounit?status.svg)](http://godoc.org/github.com/hexdigest/gounit)

GoUnit is a unit tests generator for Go programming language

The goal of the project is to generate more convenient test stubs than GoTests does and also to improve integration with text editors and IDEs.

There is a [Vim plugin](https://github.com/hexdigest/gounit-vim) that introduces the :GoUnit command which generates a test for the selected function.

## Installation

```
go get github.com/hexdigest/gounit/cmd/gounit
```

## Usage of GoUnit

```
  -all
    	generate tests for all functions (default true)
  -c string
    	comment that will be inserted into the generated test
  -f value
    	comma-separated function names to generate tests for
  -h	display this help text and exit
  -i string
    	input file name
  -json
    	read JSON-encoded input parameters from stdin
    	please see http://github.com/hexdigest/gounit for details
  -l value
    	comma-separated line numbers (starting with 1) to look for the function declarations
  -o string
    	output file name (optional)
  -stdin
    	use stdin rather than reading the input file
  -stdout
    	use stdout rather than writing to the output file
```

## JSON mode (-json command line flag)
In JSON mode GoUnit reads [JSON requests](https://github.com/hexdigest/gounit/blob/master/client.go#L5) from Stdin in a loop and produces [JSON responses](https://github.com/hexdigest/gounit/blob/master/client.go#L16) with generated test(s) that are written to Stdout.
Using this mode you can generate as many tests as you want by running GoUnit executable only once.

## Problems of GoTests

* Function name matching doesn't work if you have two methods with identical names in one file. GoTests will generate two tests, not one.
* Errors go to Stdout
* No exit codes
* TODO: comment in the generated test may change so you can't rely on it to place the carrige in the correct place after the test is generated
* You can't check for the particular error returned by the tested function with the GoTests

## GoUnit
* Takes a line number instead of the function name (in Go there can be only one function declaration on one line)
* Errors go to Stderr
* Special exit codes for input/output errors
* Special flag to read from Stdin
* Special flag to write to Stdout
* Special flag to set a //TODO comment
