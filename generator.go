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
		"join": strings.Join,
		"receiver": func() string {
			if g.Func.ReceiverType() == "" {
				return ""
			}

			return strings.Replace(g.Func.ReceiverType(), "*", "", -1) + "."
		},
		"want": func(s string) string { return strings.Replace(s, "got", "want", 1) },
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
func {{ $func.TestName }}(t *testing.T) {
	{{- if (gt $func.NumParams 0) }}
		type args struct {
			{{ range $param := $func.Params }}
				{{- $param}}
			{{ end }}
		}
	{{ end -}}
	tests := []struct {
		name string
		{{- if $func.IsMethod }}
			init func(t *testing.T) {{ $func.ReceiverType }}
			inspect func(r {{ $func.ReceiverType }}, t *testing.T) //inspects receiver after test run
		{{ end }}
		{{- if (gt $func.NumParams 0) }}
			args func(t *testing.T) args
		{{ end }}
		{{ range $result := $func.Results }}
			{{ want $result -}}
		{{ end }}
		{{- if $func.ReturnsError }}
			wantErr bool
			inspectErr func (err error, t *testing.T) //use for more precise error evaluation after test
		{{ end -}}
	}{
		{{- if eq .Comment "" }}
			//TODO: Add test cases
		{{else}}
			//{{ .Comment }}
		{{end -}}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			{{- if (gt $func.NumParams 0) }}
				tArgs := tt.args(t)
			{{ end -}}
			{{ if $func.IsMethod }}
				receiver := tt.init(t)
				{{ if (gt $func.NumResults 0) }}{{ join $func.ResultsNames ", " }} := {{end}}receiver.{{$func.Name}}(
					{{- range $i, $pn := $func.ParamsNames }}
						{{- if not (eq $i 0)}},{{end}}tArgs.{{ $pn }}{{ end }})

				if tt.inspect != nil {
					tt.inspect(receiver, t)
				}
			{{ else }}
				{{ if (gt $func.NumResults 0) }}{{ join $func.ResultsNames ", " }} := {{end}}{{$func.Name}}(
					{{- range $i, $pn := $func.ParamsNames }}
						{{- if not (eq $i 0)}},{{end}}tArgs.{{ $pn }}{{ end }})
			{{end}}
			{{ range $result := $func.ResultsNames }}
				{{ if (eq $result "err") }}
					if (err != nil) != tt.wantErr {
						t.Fatalf("{{ receiver }}{{ $func.Name }} error = %v, wantErr: %t", err, tt.wantErr)
					}

					if tt.inspectErr!= nil {
						tt.inspectErr(err, t)
					}
				{{ else }}
					if !reflect.DeepEqual({{ $result }}, tt.{{ want $result }}) {
						t.Errorf("{{ receiver }}{{ $func.Name }} {{ $result }} = %v, {{ want $result }}: %v", {{ $result }}, tt.{{ want $result }})
					}
				{{end -}}
			{{end -}}
		})
	}
}`
