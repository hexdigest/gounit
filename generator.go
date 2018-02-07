package gounit

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

//Generator is used to generate a test stub for function Func
type Generator struct {
	fs      *token.FileSet
	funcs   []*Func
	imports []*ast.ImportSpec
	pkg     string
	opt     Options
	buf     *bytes.Buffer
}

//NewGenerator returns a pointer to Generator
func NewGenerator(opt Options, src, testSrc io.Reader) (*Generator, error) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, opt.InputFile, src, 0)
	if err != nil {
		return nil, ErrFailedToParseInFile.Format(err)
	}

	visitor := NewVisitor(func(fd *ast.FuncDecl) bool {
		return opt.Lines.Include(fs.Position(fd.Pos()).Line) || opt.Functions.Include(fd.Name.Name)
	})

	ast.Walk(visitor, file)

	funcs := visitor.Funcs()

	if len(funcs) == 0 {
		return nil, ErrFuncNotFound
	}

	var buf = bytes.NewBuffer([]byte{})

	if testSrc != nil {
		tr := io.TeeReader(testSrc, buf)

		file, err := parser.ParseFile(fs, opt.OutputFile, tr, 0)
		if err != nil {
			return nil, ErrFailedToParseOutFile.Format(err)
		}

		funcs = findMissingTests(file, funcs)
	}

	return &Generator{
		buf:     buf,
		opt:     opt,
		fs:      fs,
		funcs:   funcs,
		imports: file.Imports,
		pkg:     file.Name.String(),
	}, nil
}

func (g *Generator) Write(w io.Writer) error {
	if g.buf.Len() == 0 {
		if err := g.WriteHeader(g.buf); err != nil {
			return ErrGenerateHeader.Format(err)
		}
	}

	if err := g.WriteTests(g.buf); err != nil {
		return ErrGenerateTest.Format(err)
	}

	formattedSource, err := imports.Process(g.opt.OutputFile, g.buf.Bytes(), nil)
	if err != nil {
		return ErrFixImports.Format(err)
	}

	if _, err = w.Write(formattedSource); err != nil {
		return ErrWriteTest.Format(err)
	}

	return nil
}

//WriteHeader writes a package name and imports specs to w
func (g *Generator) WriteHeader(w io.Writer) error {
	funcs := template.FuncMap{
		"ast": func(n ast.Node) string {
			return nodeToString(g.fs, n)
		},
	}
	return g.processTemplate(w, "header", headerTemplate, funcs, struct {
		Imports []*ast.ImportSpec
		Package string
	}{
		Imports: g.imports,
		Package: g.pkg,
	})
}

//WriteTest writes a test stub to the w
func (g *Generator) WriteTests(w io.Writer) error {
	helpers := template.FuncMap{
		"join": strings.Join,
		"ast": func(n ast.Node) string {
			return nodeToString(g.fs, n)
		},
		"params": func(f *Func) []string {
			return f.Params(g.fs)
		},
		"results": func(f *Func) []string {
			return f.Results(g.fs)
		},
		"receiver": func(f *Func) string {
			if f.ReceiverType() == nil {
				return ""
			}

			return strings.Replace(nodeToString(g.fs, f.ReceiverType()), "*", "", -1) + "."
		},
		"want": func(s string) string { return strings.Replace(s, "got", "want", 1) },
	}

	for _, f := range g.funcs {
		err := g.processTemplate(w, "test", testTemplate, helpers, struct {
			Func    *Func
			Comment string
		}{
			Func:    f,
			Comment: g.opt.Comment,
		})

		if err != nil {
			return fmt.Errorf("failed to write test: %v", err)
		}
	}

	return nil
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
			{{ range $param := params $func }}
				{{- $param}}
			{{ end }}
		}
	{{ end -}}
	tests := []struct {
		name string
		{{- if $func.IsMethod }}
			init func(t *testing.T) {{ ast $func.ReceiverType }}
			inspect func(r {{ ast $func.ReceiverType }}, t *testing.T) //inspects receiver after test run
		{{ end }}
		{{- if (gt $func.NumParams 0) }}
			args func(t *testing.T) args
		{{ end }}
		{{ range $result := results $func}}
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
						t.Fatalf("{{ receiver $func }}{{ $func.Name }} error = %v, wantErr: %t", err, tt.wantErr)
					}

					if tt.inspectErr!= nil {
						tt.inspectErr(err, t)
					}
				{{ else }}
					if !reflect.DeepEqual({{ $result }}, tt.{{ want $result }}) {
						t.Errorf("{{ receiver $func }}{{ $func.Name }} {{ $result }} = %v, {{ want $result }}: %v", {{ $result }}, tt.{{ want $result }})
					}
				{{end -}}
			{{end -}}
		})
	}
}`
