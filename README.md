# GoUnit [![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/hexdigest/gounit/blob/master/LICENSE) [![Build Status](https://travis-ci.org/hexdigest/gounit.svg?branch=master)](https://travis-ci.org/hexdigest/gounit) [![Coverage Status](https://coveralls.io/repos/github/hexdigest/gounit/badge.svg?branch=master)](https://coveralls.io/github/hexdigest/gounit?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hexdigest/gounit)](https://goreportcard.com/report/github.com/hexdigest/gounit) [![GoDoc](https://godoc.org/github.com/hexdigest/gounit?status.svg)](http://godoc.org/github.com/hexdigest/gounit)

GoUnit is a unit tests generator for Go programming language

Problems of GoTests
* StrictMatching doesn't work if you have two methods with identical names in one file it will generate two tests, not one
* Errors going to stdout
* No exit codes
* TODO: comment may change so you can't rely on it to place carrige to the right place after test is generated
* Can't check for particular error returned by the tested function

Proposed solution:
* Pass line number (in Go ther can be only one function declaration on one line)
* Separate generated tests (Stdout) and error messages (Stderr)
* Introduce exit codes for all error cases (FS errors, Parsing errors, command line arguments errors, etc)
* Special flag to read from Stdin
* Special flag to write to Stdout
* Special flag to set a //TODO comment
