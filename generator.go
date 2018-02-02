package gounit

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
	"text/template"
)

//Generator is used to generate a test stub for function Func
type Generator struct {
	Comment string
	Func    *Func

	fs *token.FileSet
}

//NewGenerator returns a pointer to Generator
func NewGenerator(fs *token.FileSet, fn *Func, opt Options) *Generator {
	return &Generator{
		Comment: opt.Comment,
		Func:    fn,
		fs:      fs,
	}
}

//WriteHeader writes a package name and imports specs to w
func (g *Generator) WriteHeader(w io.Writer, pkg string, imports []*ast.ImportSpec) error {
	funcs := template.FuncMap{
		"ast": func(n ast.Node) string {
			return nodeToString(g.fs, n)
		},
	}
	return g.processTemplate(w, "header", headerTemplate, funcs, map[string]interface{}{
		"Package": pkg,
		"Imports": imports,
	})
}

//WriteTest writes a test stub to the w
func (g *Generator) WriteTest(w io.Writer) error {
	funcs := template.FuncMap{
		"join":   strings.Join,
		"unstar": func(s string) string { return strings.Replace(s, "*", "", -1) },
	}
	return g.processTemplate(w, "test", testTemplate, funcs, g)
}

func (g *Generator) processTemplate(w io.Writer, tmplName, tmplBody string, funcs template.FuncMap, data interface{}) error {
	tmpl := template.New(tmplName)

	if funcs != nil {
		tmpl = tmpl.Funcs(funcs)
	}

	tmpl, err := tmpl.Parse(tmplBody)
	if err != nil {
		return fmt.Errorf("failed to parse %s template: %v", tmplName, err)
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute %s template: %v", tmplName, err)
	}

	return nil
}

var headerTemplate = `package {{.Package}}

import(
	"testing"
	"reflect"{{range $import := .Imports}}
	{{ast $import}}{{end}}
)`

var testTemplate = `{{$func := .Func}}
{{if .Func.IsMethod}}
	func {{$func.TestName}}(t *testing.T) {
		{{if (gt .Func.NumParams 0)}}type args struct { {{range $param := .Func.Params}}
			{{$param}}{{end}}
		}{{end}}

		tests := []struct {
			name string
			{{if (gt .Func.NumParams 0)}}args func(t *testing.T) args{{end}}
			init func(t *testing.T) {{.Func.ReceiverType}}
			inspect func(r {{.Func.ReceiverType}}, t *testing.T) //inspects receiver after method run
			{{range $result := $func.Results}}
			{{$result}}{{end}}
			{{if .Func.ReturnsError}}wantErr bool
				inspectErr func (err error, t *testing.T) //use for more precise error evaluation after test
			{{end}}
		}{
			{{if eq .Comment ""}}//TODO: Add test cases{{else}}//{{.Comment}}{{end}}
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				{{if (gt .Func.NumParams 0)}}tArgs := tt.args(t){{end}}
				receiver := tt.init(t)
				{{if (eq .Func.NumResults 0)}}receiver.{{.Func.Name}}({{range $pn := .Func.ParamsNames}}tc.{{$pn}},{{end}})
				{{else}}{{join .Func.ResultsNames ", "}} := receiver.{{.Func.Name}}({{range $pn := .Func.ParamsNames}}tArgs.{{$pn}},{{end}})
				{{end}}
				if tt.inspect != nil {
					tt.inspect(receiver, t)
				}
				{{range $result := .Func.ResultsNames}}
					{{if (eq $result "err")}}
						if (err != nil) != tt.wantErr {
							t.Fatalf("{{unstar $func.ReceiverType}}.{{$func.Name}} error = %v, wantErr: %t", err, tt.wantErr)
						}

						if tt.inspectErr!= nil {
							tt.inspectErr(err, t)
						}
					{{else}}
						if !reflect.DeepEqual({{$result}}, tt.{{$result}}) {
							t.Errorf("{{unstar $func.ReceiverType}}.{{$func.Name}} {{$result}} = %v, {{$result}}: %v", {{$result}}, tt.{{$result}})
						}
					{{end}}
				{{end}}
			})
		}
	}
{{else}}
	func {{$func.TestName}}(t *testing.T) {
		{{if (gt .Func.NumParams 0)}}type args struct { {{range $param := .Func.Params}}
			{{$param}}{{end}}
		}{{end}}

		tests := []struct {
			name string
			{{if (gt .Func.NumParams 0)}}args func(t *testing.T) args{{end}}
			{{range $result := $func.Results}}
			{{$result}}{{end}}
			{{if .Func.ReturnsError}}wantErr bool
				inspectErr func (err error, t *testing.T) //use for more precise error evaluation after test
			{{end}}
		}{
			{{if eq .Comment ""}}//TODO: Add test cases{{else}}//{{.Comment}}{{end}}
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				{{if (gt .Func.NumParams 0)}}tArgs := tt.args(t){{end}}
				{{join .Func.ResultsNames ", "}} := {{.Func.Name}}({{range $pn := .Func.ParamsNames}}tArgs.{{$pn}},{{end}})
				{{range $result := .Func.ResultsNames}}
					{{if (eq $result "err")}}
						if (err != nil) != tt.wantErr {
							t.Fatalf("{{$func.Name}} error = %v, wantErr: %t", err, tt.wantErr)
						}

						if tt.inspectErr!= nil {
							tt.inspectErr(err, t)
						}
					{{else}}
						if !reflect.DeepEqual({{$result}}, tt.{{$result}}) {
							t.Errorf("{{$func.Name}} {{$result}} = %v, {{$result}}: %v", {{$result}}, tt.{{$result}})
						}
					{{end}}
				{{end}}
			})
		}
	}
{{end}}`
