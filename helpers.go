package gounit

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"text/template"
)

type matchFunc func(*ast.FuncDecl) bool

//findMissingTests filters funcs slice and returns only those functions that don't have tests yet
func findMissingTests(file *ast.File, funcs []*Func) []*Func {
	tests := findFunctions(file.Decls, func(fd *ast.FuncDecl) bool {
		for _, sourceFunc := range funcs {
			f := NewFunc(fd)
			if f.ReceiverType() == nil && f.Name() == sourceFunc.TestName() {
				return true
			}
		}
		return false
	})

	dontHaveTests := []*Func{}
	for _, f := range funcs {
		testIsFound := false
		for _, test := range tests {
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

//findFunctions finds all matching function declarations
func findFunctions(decls []ast.Decl, match matchFunc) []*Func {
	var funcs []*Func

	for _, decl := range decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && match(fd) {
			funcs = append(funcs, NewFunc(fd))
		}
	}

	return funcs
}

//nodeToString returns a string representation of an AST node
//as it has in the original source code
func nodeToString(fs *token.FileSet, n ast.Node) string {
	b := bytes.NewBuffer([]byte{})
	printer.Fprint(b, fs, n)
	return b.String()
}

//templateHelpers return FuncMap of template helpers to use within a template
func templateHelpers(fs *token.FileSet) template.FuncMap {
	return template.FuncMap{
		"ast": func(n ast.Node) string {
			return nodeToString(fs, n)
		},
		"join": strings.Join,
		"params": func(f *Func) []string {
			return f.Params(fs)
		},
		"results": func(f *Func) []string {
			return f.Results(fs)
		},
		"receiver": func(f *Func) string {
			if f.ReceiverType() == nil {
				return ""
			}

			return strings.Replace(nodeToString(fs, f.ReceiverType()), "*", "", -1) + "."
		},
		"want": func(s string) string { return strings.Replace(s, "got", "want", 1) },
	}
}
