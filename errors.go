package gounit

import "fmt"

const (
	//standard exit codes
	ExitCodeOK             = 0
	ExitCodeErrGeneric     = 1
	ExitCodeErrCommandLine = 2
)

var (
	ErrGenerateHeader        = NewError(3, "failed to write header: %v")
	ErrGenerateTest          = NewError(4, "failed to write test: %v")
	ErrFuncNotFound          = NewError(6, "unable to find a function declaration")
	ErrFailedToParseInFile   = NewError(9, "failed to parse input file: %v")
	ErrFailedToParseOutFile  = NewError(10, "failed to parse output file: %v")
	ErrFailedToOpenInFile    = NewError(11, "failed to open input file: %v")
	ErrFailedToOpenOutFile   = NewError(12, "failed to open output file: %v")
	ErrFailedToCreateOutFile = NewError(13, "failed to create output file: %v")
	ErrInputFileDoesNotExist = NewError(14, "input file does not exist")
	ErrSeekFailed            = NewError(16, "failed to seek: %v")
	ErrFixImports            = NewError(17, "failed to fix imports: %v")
	ErrWriteTest             = NewError(18, "failed to write generated test: %v")
)

type Error struct {
	code   int
	format string
}

func NewError(code int, format string) *Error {
	return &Error{code: code, format: format}
}

func (e *Error) Format(args ...interface{}) *Error {
	return NewError(e.code, fmt.Sprintf(e.format, args...))
}

func (e *Error) Error() string {
	return e.format
}

func (e *Error) Code() int {
	return e.code
}
