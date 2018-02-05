package gounit

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
)

//FindSourceFunc returns *Func struct for the function declaration on
//line opt.LineNumber
func FindSourceFunc(fs *token.FileSet, file *ast.File, opt Options) (*Func, error) {
	if file.Name == nil {
		return nil, errors.New("input file doesn't contain package name")
	}

	if fs.File(token.Pos(1)).LineCount() < opt.LineNumber {
		return nil, fmt.Errorf("line number is too big: %d", opt.LineNumber)
	}

	visitor := NewVisitor(func(fd *ast.FuncDecl) bool {
		return fs.Position(fd.Pos()).Line == opt.LineNumber || fd.Name.Name == opt.Function
	})

	ast.Walk(visitor, file)

	foundFunc := visitor.Func()
	if foundFunc == nil {
		return nil, errors.New("unable to find a function declaration on the given line")
	}

	return NewFunc(fs, foundFunc), nil
}

//IsTestExist checks if the test for function fn
//is already exist in the file represented by r
func IsTestExist(fs *token.FileSet, r io.Reader, fn *Func, opt Options) (bool, error) {
	file, err := parser.ParseFile(fs, opt.InputFile, r, 0)
	if err != nil {
		return false, fmt.Errorf("failed to parse file: %v", err)
	}

	visitor := NewVisitor(func(fd *ast.FuncDecl) bool {
		return fd.Name.Name == fn.TestName()
	})

	ast.Walk(visitor, file)

	return visitor.Func() != nil, nil
}

//nodeToString returns a string representation of an AST node
//as it has in the original source code
func nodeToString(fs *token.FileSet, n ast.Node) string {
	b := bytes.NewBuffer([]byte{})
	printer.Fprint(b, fs, n)
	return b.String()
}
