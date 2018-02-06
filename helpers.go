package gounit

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
)

//findMissingTests filters funcs slice and returns only those functions that don't have tests yet
func findMissingTests(file *ast.File, funcs []*Func) []*Func {
	visitor := NewVisitor(func(fd *ast.FuncDecl) bool {
		for _, sourceFunc := range funcs {
			f := NewFunc(fd)
			if f.ReceiverType() == nil && f.Name() == sourceFunc.TestName() {
				return true
			}
		}
		return false
	})

	ast.Walk(visitor, file)

	dontHaveTests := []*Func{}
	for _, f := range funcs {
		testIsFound := false
		for _, test := range visitor.Funcs() {
			if test.Name() == f.TestName() {
				testIsFound = true
				break
			}
		}
		if !testIsFound {
			dontHaveTests = append(dontHaveTests, f)
		}
	}

	return dontHaveTests
}

//nodeToString returns a string representation of an AST node
//as it has in the original source code
func nodeToString(fs *token.FileSet, n ast.Node) string {
	b := bytes.NewBuffer([]byte{})
	printer.Fprint(b, fs, n)
	return b.String()
}
