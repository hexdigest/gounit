package gounit

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

//Func is a wrapper around ast.FuncDecl containing few methods
//to use within a test template
type Func struct {
	Signature *ast.FuncDecl
	fs        *token.FileSet
}

//NewFunc returns pointer to the Func struct
func NewFunc(fs *token.FileSet, sig *ast.FuncDecl) *Func {
	return &Func{Signature: sig, fs: fs}
}

//NumParams returns a number of the function params
func (f *Func) NumParams() int {
	return f.Signature.Type.Params.NumFields()
}

//NumResults returns a number of the function results
func (f *Func) NumResults() int {
	if f.Signature.Type.Results == nil {
		return 0
	}
	return f.Signature.Type.Results.NumFields()
}

//Params returns a list of the function params with their types
func (f *Func) Params() []string {
	if f.Signature.Type.Params == nil {
		return nil
	}

	params := []string{}
	for i, p := range f.Signature.Type.Params.List {
		for _, n := range p.Names {
			param := n.Name
			if i == len(f.Signature.Type.Params.List)-1 && f.IsVariadic() {
				param += " []" + nodeToString(f.fs, p.Type.(*ast.Ellipsis).Elt)
			} else {
				param += " " + nodeToString(f.fs, p.Type)
			}

			params = append(params, param)
		}
	}

	return params
}

//Results returns a list of the function results with their types
//if function's last param is an error it is not included in the result slice
func (f *Func) Results() []string {
	if f.Signature.Type.Results == nil {
		return nil
	}

	var (
		results []string
		n       = 1
	)
	for _, r := range f.Signature.Type.Results.List {
		if len(r.Names) > 0 {
			for range r.Names {
				results = append(results, fmt.Sprintf("got%d %s", n, nodeToString(f.fs, r.Type)))
				n++
			}
		} else {
			results = append(results, fmt.Sprintf("got%d %s", n, nodeToString(f.fs, r.Type)))
			n++
		}
	}

	if f.ReturnsError() {
		results = results[:len(results)-1]
	}

	return results
}

//ParamsNames returns a list of the function params' names
func (f *Func) ParamsNames() []string {
	if f.Signature.Type.Params == nil {
		return nil
	}

	names := []string{}
	for i, p := range f.Signature.Type.Params.List {
		for _, n := range p.Names {
			name := n.Name
			if i == f.NumParams()-1 && f.IsVariadic() {
				name += "..."
			}
			names = append(names, name)
		}
	}

	return names
}

//ResultsNames returns a list of the function results' names
//if function's last result is an error the name of param is "err"
func (f *Func) ResultsNames() []string {
	if f.Signature.Type.Results == nil {
		return nil
	}

	var (
		names []string
		n     = 1
	)
	for _, r := range f.Signature.Type.Results.List {
		if len(r.Names) > 0 {
			for range r.Names {
				names = append(names, fmt.Sprintf("got%d", n))
				n++
			}
		} else {
			names = append(names, fmt.Sprintf("got%d", n))
			n++
		}
	}

	if f.ReturnsError() {
		names[len(names)-1] = "err"
	}

	return names
}

//TestName returns a name of the test
func (f *Func) TestName() string {
	name := "Test_"
	if f.IsMethod() {
		name += strings.Replace(f.ReceiverType(), "*", "", -1) + "_"
	}

	return name + f.Signature.Name.String()
}

//IsMethod returns true if the function is a method
func (f *Func) IsMethod() bool {
	return f.Signature.Recv != nil
}

//ReceiverType returns a type of the method receiver
func (f *Func) ReceiverType() string {
	if f.Signature.Recv == nil {
		return ""
	}

	return nodeToString(f.fs, f.Signature.Recv.List[0].Type)
}

//ReturnsError returns true if the function's last param's type is error
func (f *Func) ReturnsError() bool {
	lastResult := f.LastResult()
	if lastResult == nil {
		return false
	}

	ident, ok := lastResult.Type.(*ast.Ident)
	return ok && ident.Name == "error"
}

//LastParam returns function's last param
func (f *Func) LastParam() *ast.Field {
	numFields := len(f.Signature.Type.Params.List)
	if numFields == 0 {
		return nil
	}

	return f.Signature.Type.Params.List[numFields-1]
}

//LastResult returns function's last result
func (f *Func) LastResult() *ast.Field {
	if f.Signature.Type.Results == nil {
		return nil
	}

	numFields := len(f.Signature.Type.Results.List)
	if numFields == 0 {
		return nil
	}

	return f.Signature.Type.Results.List[numFields-1]
}

//IsVariadic returns true if it's the variadic function
func (f *Func) IsVariadic() bool {
	lastParam := f.LastParam()
	if lastParam == nil {
		return false
	}

	_, isVariadic := lastParam.Type.(*ast.Ellipsis)

	return isVariadic
}
