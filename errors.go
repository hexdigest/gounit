package gounit

import "fmt"

type GenericError string

func (e GenericError) Format(args ...interface{}) GenericError {
	return GenericError(fmt.Sprintf(string(e), args...))
}

func (e GenericError) Error() string {
	return string(e)
}

type CommandLineError string

func (e CommandLineError) Error() string {
	return string(e)
}
