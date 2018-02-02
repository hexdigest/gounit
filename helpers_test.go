package gounit

import (
	"go/ast"
	"go/token"
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_IsTestExist(t *testing.T) {
	type args struct {
		fs  *token.FileSet
		r   io.Reader
		fn  *Func
		opt Options
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		got1       bool
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test

	}{
		{
			name: "failed to parse file",
			args: func(*testing.T) args {
				return args{
					r:  strings.NewReader(""),
					fs: token.NewFileSet(),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if !strings.HasPrefix(err.Error(), "failed to parse file:") {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "test does not exist",
			args: func(*testing.T) args {
				return args{
					r:  strings.NewReader("package test\n"),
					fs: token.NewFileSet(),
					fn: &Func{Signature: &ast.FuncDecl{}},
				}
			},
			wantErr: false,
		},
		{
			name: "test exists",
			args: func(*testing.T) args {
				return args{
					r:  strings.NewReader("package test\nfunc Test_test() {\n}"),
					fs: token.NewFileSet(),
					fn: &Func{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "test"}}},
				}
			},
			got1:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1, err := IsTestExist(tArgs.fs, tArgs.r, tArgs.fn, tArgs.opt)

			if !reflect.DeepEqual(got1, tt.got1) {
				t.Errorf("IsTestExist got1 = %v, got1: %v", got1, tt.got1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("IsTestExist error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}

		})
	}
}

func Test_FindSourceFunc(t *testing.T) {
	type args struct {
		fs   *token.FileSet
		file *ast.File
		opt  Options
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		got1       *Func
		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test

	}{
		{
			name: "no package name",
			args: func(*testing.T) args {
				return args{
					file: &ast.File{},
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if err.Error() != "input file doesn't contain package name" {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "line number is too big",
			args: func(*testing.T) args {
				fs := token.NewFileSet()
				fs.AddFile("fake.go", 1, 10)
				return args{
					file: &ast.File{Name: &ast.Ident{}},
					opt:  Options{LineNumber: 100},
					fs:   fs,
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if err.Error() != "line number is too big: 100" {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "func not found",
			args: func(*testing.T) args {
				fs := token.NewFileSet()
				file := fs.AddFile("fake.go", 1, 1000)
				file.AddLine(1)
				return args{
					file: &ast.File{Name: &ast.Ident{}},
					opt:  Options{LineNumber: 1},
					fs:   fs,
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				if err.Error() != "unable to find a function declaration on the given line" {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1, err := FindSourceFunc(tArgs.fs, tArgs.file, tArgs.opt)

			if !reflect.DeepEqual(got1, tt.got1) {
				t.Errorf("FindSourceFunc got1 = %v, got1: %v", got1, tt.got1)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindSourceFunc error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
