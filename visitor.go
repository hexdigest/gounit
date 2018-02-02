package gounit

import (
	"go/ast"
)

type matchFunc func(*ast.FuncDecl) bool

//Visitor finds first function declaration that matches the match func
type Visitor struct {
	found *ast.FuncDecl
	match matchFunc
}

//NewVisitor returns a pointer to the Visitor struct
func NewVisitor(match matchFunc) *Visitor {
	return &Visitor{match: match}
}

// Visit implements ast.Visitor interface
func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if v.found != nil {
		return nil
	}

	if fd, ok := node.(*ast.FuncDecl); ok {
		if v.match(fd) {
			v.found = fd
			return nil
		}
	}

	return v
}

//Func returns found function declaration
func (v *Visitor) Func() *ast.FuncDecl {
	return v.found
}
