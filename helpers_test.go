package gounit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func Test_nodeToString(t *testing.T) {
	type args struct {
		fs *token.FileSet
		n  ast.Node
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 string
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{
					fs: token.NewFileSet(),
					n:  &ast.Ident{Name: "node"},
				}
			},
			want1: "node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1 := nodeToString(tArgs.fs, tArgs.n)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("nodeToString got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func Test_findMissingTests(t *testing.T) {
	const gofile = `package gofile

	func Test_function() error {
		return nil
	}`

	type args struct {
		file  *ast.File
		funcs []*Func
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 []*Func
	}{
		{
			name: "no functions",
			args: func(*testing.T) args {
				return args{
					file: &ast.File{},
				}
			},
			want1: []*Func{},
		},
		{
			name: "tests not found",
			args: func(t *testing.T) args {
				fs := token.NewFileSet()
				file, err := parser.ParseFile(fs, "file.go", []byte(gofile), 0)
				if err != nil {
					t.Fatalf("failed to parse file: %v", err)
				}
				return args{
					file:  file,
					funcs: []*Func{{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "missing_function"}}}},
				}
			},
			want1: []*Func{{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "missing_function"}}}},
		},
		{
			name: "tests found",
			args: func(t *testing.T) args {
				fs := token.NewFileSet()
				file, err := parser.ParseFile(fs, "file.go", []byte(gofile), 0)
				if err != nil {
					t.Fatalf("failed to parse file: %v", err)
				}
				return args{
					file:  file,
					funcs: []*Func{{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "function"}}}},
				}
			},
			want1: []*Func{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := findMissingTests(tArgs.file, tArgs.funcs)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("findMissingTests got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func Test_templateHelpers(t *testing.T) {
	helpers := templateHelpers(token.NewFileSet())

	astHelper, ok := helpers["ast"].(func(ast.Node) string)
	if !ok {
		t.Fatalf("unexpected ast helper type")
	}

	s := astHelper(nil)
	if s != "" {
		t.Errorf("unexpected ast helper result: %s", s)
	}

	paramsHelper, ok := helpers["params"].(func(*Func) []string)
	if !ok {
		t.Fatalf("unexpected params helper type")
	}

	params := paramsHelper(&Func{Signature: &ast.FuncDecl{Type: &ast.FuncType{Params: &ast.FieldList{}}}})
	if len(params) > 0 {
		t.Errorf("unexpected params len: %d", len(params))
	}

	resultsHelper, ok := helpers["results"].(func(*Func) []string)
	if !ok {
		t.Fatalf("unexpected results helper type")
	}

	results := resultsHelper(&Func{Signature: &ast.FuncDecl{Type: &ast.FuncType{}}})
	if len(results) > 0 {
		t.Errorf("unexpected params len: %d", len(results))
	}

	receiverHelper, ok := helpers["receiver"].(func(*Func) string)
	if !ok {
		t.Fatalf("unexpected receiver helper type")
	}

	noReceiver := receiverHelper(&Func{Signature: &ast.FuncDecl{}})
	if noReceiver != "" {
		t.Errorf("unexpected receiver: %s", noReceiver)
	}

	receiver := receiverHelper(&Func{Signature: &ast.FuncDecl{Recv: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "Receiver"}}}}}})
	if receiver != "Receiver." {
		t.Errorf("unexpected receiver: %s", receiver)
	}

	wantHelper, ok := helpers["want"].(func(string) string)
	if !ok {
		t.Fatalf("unexpected want helper type")
	}

	want := wantHelper("got1")
	if want != "want1" {
		t.Errorf("unexpected want helper result: %s", want)
	}
}
