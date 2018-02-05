package gounit

import (
	"go/ast"
	"go/token"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

type expectWriter struct {
	t        *testing.T
	expected []byte
	written  []byte
}

func newExpectPrefixWriter(t *testing.T, s string) *expectWriter {
	return &expectWriter{t: t, expected: []byte(s)}
}

func (ew *expectWriter) Write(p []byte) (int, error) {
	ew.written = append(ew.written, p...)
	var prefixMatch bool
	if len(ew.expected) > len(ew.written) {
		prefixMatch = strings.HasPrefix(string(ew.expected), string(ew.written))
	} else {
		prefixMatch = strings.HasPrefix(string(ew.written), string(ew.expected))
	}

	if !prefixMatch {
		ew.t.Fatalf("unexpected argument, got: %q, want: %q", string(ew.written), string(ew.expected))
	}

	return len(p), nil
}

type errorWriter struct {
	err error
}

func (ew errorWriter) Write([]byte) (int, error) {
	return 0, ew.err
}

func Test_Generator_processTemplate(t *testing.T) {
	type args struct {
		w        io.Writer
		tmplName string
		tmplBody string
		funcs    template.FuncMap
		data     interface{}
	}

	tests := []struct {
		name    string
		args    func(t *testing.T) args
		init    func(t *testing.T) *Generator
		inspect func(r *Generator, t *testing.T) //inspects receiver after method run

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test

	}{
		{
			name: "template parsing failed",
			args: func(t *testing.T) args {
				return args{
					tmplName: "test",
					tmplBody: "{{.",
				}
			},
			init:    func(*testing.T) *Generator { return &Generator{} },
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if !strings.HasPrefix(err.Error(), "failed to parse test template:") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "template execution failed",
			args: func(t *testing.T) args {
				return args{
					tmplName: "test",
					tmplBody: "{{.}}",
					w:        errorWriter{io.EOF},
				}
			},
			init:    func(*testing.T) *Generator { return &Generator{} },
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if !strings.HasPrefix(err.Error(), "failed to execute test template:") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "success",
			args: func(t *testing.T) args {
				return args{
					tmplName: "test",
					tmplBody: "{{.}}",
					w:        ioutil.Discard,
				}
			},
			init:    func(*testing.T) *Generator { return &Generator{} },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			err := receiver.processTemplate(tArgs.w, tArgs.tmplName, tArgs.tmplBody, tArgs.funcs, tArgs.data)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generator.processTemplate error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}

		})
	}
}

func Test_Generator_WriteHeader(t *testing.T) {
	type args struct {
		w       io.Writer
		pkg     string
		imports []*ast.ImportSpec
	}

	tests := []struct {
		name    string
		args    func(t *testing.T) args
		init    func(t *testing.T) *Generator
		inspect func(r *Generator, t *testing.T) //inspects receiver after method run

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test

	}{
		{
			name: "bad writer",
			args: func(t *testing.T) args {
				return args{
					w: errorWriter{io.EOF},
				}
			},
			init:    func(t *testing.T) *Generator { return &Generator{} },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			err := receiver.WriteHeader(tArgs.w, tArgs.pkg, tArgs.imports)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generator.WriteHeader error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}

		})
	}
}

func Test_Generator_WriteTest(t *testing.T) {
	type args struct {
		w io.Writer
	}

	tests := []struct {
		name    string
		args    func(t *testing.T) args
		init    func(t *testing.T) *Generator
		inspect func(r *Generator, t *testing.T) //inspects receiver after method run

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test

	}{
		{
			name: "bad writer",
			args: func(t *testing.T) args {
				return args{
					w: errorWriter{io.EOF},
				}
			},
			init:    func(t *testing.T) *Generator { return &Generator{} },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			err := receiver.WriteTest(tArgs.w)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generator.WriteTest error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}

		})
	}
}

func Test_NewGenerator(t *testing.T) {
	type args struct {
		fs  *token.FileSet
		fn  *Func
		opt Options
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *Generator
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{
					fs:  token.NewFileSet(),
					fn:  &Func{},
					opt: Options{Comment: "TODO:"},
				}
			},
			want1: &Generator{
				fs:      token.NewFileSet(),
				Func:    &Func{},
				Comment: "TODO:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1 := NewGenerator(tArgs.fs, tArgs.fn, tArgs.opt)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewGenerator got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
