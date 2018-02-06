package gounit

import "fmt"

const (
	//standard exit codes
	ExitCodeOK             = 0
	ExitCodeErrCommandLine = 2
)

var (
	ErrGenerateHeader         = NewError(3, "failed to write header: %v")
	ErrGenerateTest           = NewError(4, "failed to write test: %v")
	ErrFailedToFindSourceFunc = NewError(5, "failed to find source function(s): %v")
	ErrFuncNotFound           = NewError(6, "unable to find a function declaration")
	ErrNoPackageName          = NewError(7, "input file doesn't contain package name")
	ErrLineNumberIsTooBig     = NewError(8, "line number is too big: %d")
	ErrFailedToParseInFile    = NewError(9, "failed to parse input file: %v")
	ErrFailedToParseOutFile   = NewError(10, "failed to parse output file: %v")
	ErrFailedToOpenInFile     = NewError(11, "failed to open input file: %v")
	ErrFailedToOpenOutFile    = NewError(12, "failed to open output file: %v")
	ErrFailedToCreateOutFile  = NewError(13, "failed to create output file: %v")
	ErrInputFileDoesNotExist  = NewError(14, "input file does not exist")
	ErrTestIsAlreadyExist     = NewError(15, "test is already exist")
	ErrSeekFailed             = NewError(16, "failed to seek: %v")
	ErrFixImports             = NewError(17, "failed to fix imports: %v")
	ErrWriteTest              = NewError(18, "failed to write generated test: %v")
)

type Error struct {
	code   int
	format string
}

func NewError(code int, format string) Error {
	return Error{code: code, format: format}
}

func (e Error) Format(args ...interface{}) error {
	return fmt.Errorf(e.format, args...)
}

func (e Error) Code() int {
	return e.code
}
