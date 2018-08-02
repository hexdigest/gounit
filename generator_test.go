package gounit

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
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

type errorReader struct {
	err error
}

func (er errorReader) Read([]byte) (int, error) {
	return 0, er.err
}

func TestGenerator_WriteHeader(t *testing.T) {
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
			init: func(t *testing.T) *Generator {
				return &Generator{
					headerTemplate: template.Must(template.New("test").Parse("{{.}}")),
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			err := receiver.WriteHeader(tArgs.w)

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

func TestGenerator_WriteTests(t *testing.T) {
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
			init: func(t *testing.T) *Generator {
				return &Generator{
					testTemplate: template.Must(template.New("test").Parse("success")),
					funcs:        []*Func{{}},
				}
			},
			wantErr: true,
		},
		{
			name: "success",
			args: func(t *testing.T) args {
				return args{
					w: ioutil.Discard,
				}
			},
			init: func(t *testing.T) *Generator {
				return &Generator{
					testTemplate: template.Must(template.New("test").Parse("success")),
					funcs:        []*Func{{}},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			err := receiver.WriteTests(tArgs.w)

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

func TestNewGenerator(t *testing.T) {
	type args struct {
		opt     Options
		src     io.Reader
		testSrc io.Reader
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "failed to parse input file",
			args: func(*testing.T) args {
				return args{
					src: strings.NewReader(``),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error: %v", err)
				}

				if gErr != ErrFailedToParseInFile.Format("1:1: expected 'package', found 'EOF'") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "no funcs found",
			args: func(*testing.T) args {
				return args{
					src: strings.NewReader(`package nofuncs`),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if err != ErrFuncNotFound {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "funcs found and no destination source",
			args: func(*testing.T) args {
				return args{
					opt: Options{
						Functions: []string{"function"},
					},
					src: strings.NewReader(`package nofuncs
					 func function() int {
					 	 return 0
					 }`),
				}
			},
		},
		{
			name: "funcs found and destination source is broken",
			args: func(*testing.T) args {
				return args{
					opt: Options{
						Functions: []string{"function"},
					},
					src: strings.NewReader(`package nofuncs
					 func function() int {
					 	 return 0
					 }`),
					testSrc: strings.NewReader(``),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error: %v", err)
				}

				if gErr != ErrFailedToParseOutFile.Format("1:1: expected 'package', found 'EOF'") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "funcs found and destination source is fine",
			args: func(*testing.T) args {
				return args{
					opt: Options{
						Functions: []string{"function"},
					},
					src: strings.NewReader(`package nofuncs
					 func function() int {
					 	 return 0
					 }`),
					testSrc: strings.NewReader(`package nofuncs`),
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			_, err := NewGenerator(tArgs.opt, tArgs.src, tArgs.testSrc)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewGenerator error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestGenerator_Write(t *testing.T) {
	type args struct {
		w io.Writer
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *Generator
		inspect func(r *Generator, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "failed to write header",
			init: func(t *testing.T) *Generator {
				return &Generator{
					buf:   bytes.NewBuffer([]byte{}),
					funcs: []*Func{{}},
					headerTemplate: template.Must(template.New("header").Funcs(template.FuncMap{
						"error": func() (string, error) {
							return "", errors.New("error")
						},
					}).Parse("{{ error }}")),
				}
			},
			args: func(*testing.T) args {
				return args{}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error type: %T", err)
				}

				if gErr != ErrGenerateHeader.Format("template: header:1:3: executing \"header\" at <error>: error calling error: error") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "failed to write tests",
			init: func(t *testing.T) *Generator {
				return &Generator{
					buf:            bytes.NewBuffer([]byte{}),
					funcs:          []*Func{{}},
					headerTemplate: template.Must(template.New("header").Parse("package header")),
					testTemplate: template.Must(template.New("test").Funcs(template.FuncMap{
						"error": func() (string, error) {
							return "", errors.New("error")
						},
					}).Parse("{{ error }}")),
				}
			},
			args: func(*testing.T) args {
				return args{}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error type: %T", err)
				}

				if gErr != ErrGenerateTest.Format("failed to write test: template: test:1:3: executing \"test\" at <error>: error calling error: error") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "failed to fix imports",
			init: func(t *testing.T) *Generator {
				return &Generator{
					buf:            bytes.NewBuffer([]byte{}),
					funcs:          []*Func{{}},
					headerTemplate: template.Must(template.New("header").Parse("invalid go file header")),
					testTemplate:   template.Must(template.New("test").Parse("{{ . }}")),
				}
			},
			args: func(*testing.T) args {
				return args{}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error type: %T", err)
				}

				if gErr != ErrFixImports.Format("1:1: expected 'package', found invalid") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "failed to fix imports",
			init: func(t *testing.T) *Generator {
				return &Generator{
					buf:            bytes.NewBuffer([]byte{}),
					funcs:          []*Func{{}},
					headerTemplate: template.Must(template.New("header").Parse("package header")),
					testTemplate:   template.Must(template.New("test").Parse("//comment")),
				}
			},
			args: func(*testing.T) args {
				return args{w: errorWriter{errors.New("write error")}}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				gErr, ok := err.(GenericError)
				if !ok {
					t.Fatalf("unexpected error type: %T", err)
				}

				if gErr != ErrWriteTest.Format("write error") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "success",
			init: func(t *testing.T) *Generator {
				return &Generator{
					buf:            bytes.NewBuffer([]byte{}),
					funcs:          []*Func{{}},
					headerTemplate: template.Must(template.New("header").Parse("package header")),
					testTemplate:   template.Must(template.New("test").Parse("//comment")),
				}
			},
			args: func(*testing.T) args {
				return args{w: ioutil.Discard}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.Write(tArgs.w)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generator.Write error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
