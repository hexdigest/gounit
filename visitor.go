package gounit

import (
	"go/ast"
)

type matchFunc func(*ast.FuncDecl) bool

//Visitor finds all matching func declarations
type Visitor struct {
	found []*Func
	match matchFunc
}

//NewVisitor returns a pointer to the Visitor struct
func NewVisitor(match matchFunc) *Visitor {
	return &Visitor{match: match}
}

//Visit implements ast.Visitor interface
func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if fd, ok := node.(*ast.FuncDecl); ok {
		if v.match(fd) {
			v.found = append(v.found, NewFunc(fd))
		}
	}

	return v
}

//Func returns found functions
func (v *Visitor) Funcs() []*Func {
	return v.found
}
