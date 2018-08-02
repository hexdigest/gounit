# GoUnit [![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](https://github.com/hexdigest/gounit/blob/master/LICENSE) [![Build Status](https://travis-ci.org/hexdigest/gounit.svg?branch=master)](https://travis-ci.org/hexdigest/gounit) [![Coverage Status](https://coveralls.io/repos/github/hexdigest/gounit/badge.svg?branch=master)](https://coveralls.io/github/hexdigest/gounit?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hexdigest/gounit)](https://goreportcard.com/report/github.com/hexdigest/gounit) [![GoDoc](https://godoc.org/github.com/hexdigest/gounit?status.svg)](http://godoc.org/github.com/hexdigest/gounit)

GoUnit is a unit tests generator for Go programming language

The goal of the project is to generate convenient test stubs and to improve integration with text editors and IDEs.

There are plugins for
* [Vim plugin](https://github.com/hexdigest/gounit-vim) that introduces the :GoUnit command which generates a test for the selected function.
* [Emacs](https://github.com/hexdigest/GoUnit-Emacs)
* [Atom](https://github.com/hexdigest/atom-gounit)
* [Sublime](https://github.com/hexdigest/gounit-sublime)

## Installation

```
go get -u github.com/hexdigest/gounit/cmd/gounit
```

## Usage of GoUnit

This will generate test stubs for all functions and method in file.go

```
  $ gounit gen -i file.go 
```

Run
```
  $ gounit help
```

for more options

## Custom test templates

If you're not satisfied with the code produced by the default GoUnit test template you can always write your own!
You can use [minimock](https://github.com/hexdigest/gounit/blob/master/templates/minimock) template as an example.
Here is now to switch to the custom template:

```
  $ curl https://raw.githubusercontent.com/hexdigest/gounit/master/templates/minimock > minimock
  $ gounit template add -f minimock
  $ gounit template list

    gounit templates installed

         1. standard preinstalled template
      => 2. minimock

  $ gounit template use -n 2
```

## Integration with editors and IDEs

To ease an integration of GoUnit with IDEs "gen" subcommand has a "-json" flag.
When -json flag is passed GoUnit reads [JSON requests](https://github.com/hexdigest/gounit/blob/master/client.go#L5) from Stdin in a loop and produces [JSON responses](https://github.com/hexdigest/gounit/blob/master/client.go#L16) with generated test(s) that are written to Stdout.
Using this mode you can generate as many tests as you want by running GoUnit executable only once.
