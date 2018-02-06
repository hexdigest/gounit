package gounit

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestNewFunc(t *testing.T) {
	type args struct {
		fs  *token.FileSet
		sig *ast.FuncDecl
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *Func
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{
					fs:  token.NewFileSet(),
					sig: &ast.FuncDecl{},
				}
			},
			want1: &Func{
				Signature: &ast.FuncDecl{},
				fs:        token.NewFileSet(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1 := NewFunc(tArgs.fs, tArgs.sig)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewFunc got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_NumParams(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 int
	}{
		{
			name: "success",
			init: func(*testing.T) *Func {
				fd := &ast.FuncDecl{
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
					},
				}
				return &Func{
					Signature: fd,
				}
			},
			want1: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.NumParams()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.NumParams got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_NumResults(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 int
	}{
		{
			name: "no results",
			init: func(*testing.T) *Func {
				fd := &ast.FuncDecl{
					Type: &ast.FuncType{},
				}
				return &Func{
					Signature: fd,
				}
			},
			want1: 0,
		},
		{
			name: "has results",
			init: func(*testing.T) *Func {
				fd := &ast.FuncDecl{
					Type: &ast.FuncType{
						Results: &ast.FieldList{List: []*ast.Field{{}, {}}},
					},
				}
				return &Func{
					Signature: fd,
				}
			},
			want1: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.NumResults()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.NumResults got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_Params(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 []string
	}{
		{
			name: "no params",
			init: func(*testing.T) *Func {
				return &Func{
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{},
					},
				}
			},
		},
		{
			name: "params with names",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "int"}}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
			want1: []string{"s1 string", "s2 string", "i3 []int"},
		},
		{
			name: "anonymous params",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "int"}},
							},
						}},
					},
				}
			},
			want1: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.Params()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.Params got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_Results(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 []string
	}{
		{
			name: "no results",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Type: &ast.FuncType{}}}
			},
		},
		{
			name: "results with names",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
			want1: []string{"got1 string", "got2 string", "got3 int"},
		},
		{
			name: "anonymous results",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "int"}},
							},
						}},
					},
				}
			},
			want1: []string{"got1 string", "got2 int"},
		},
		{
			name: "returns error",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "error"}},
							},
						}},
					},
				}
			},
			want1: []string{"got1 string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.Results()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.Results got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestFunc_ParamsNames(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 []string
	}{
		{
			name: "no params",
			init: func(*testing.T) *Func {
				return &Func{
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{},
					},
				}
			},
		},
		{
			name: "params with names",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "int"}}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
			want1: []string{"s1", "s2", "i3..."},
		},
		{
			name: "anonymous params",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "int"}},
							},
						}},
					},
				}
			},
			want1: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.ParamsNames()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.ParamsNames got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_ResultsNames(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 []string
	}{
		{
			name: "no results",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Type: &ast.FuncType{}}}
			},
		},
		{
			name: "results with names",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
			want1: []string{"got1", "got2", "got3"},
		},
		{
			name: "anonymous results",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "int"}},
							},
						}},
					},
				}
			},
			want1: []string{"got1", "got2"},
		},
		{
			name: "returns error",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "error"}},
							},
						}},
					},
				}
			},
			want1: []string{"got1", "err"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.ResultsNames()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.ResultsNames got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_TestName(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 string
	}{
		{
			name: "func",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "func"}}}
			},
			want1: "Test_func",
		},
		{
			name: "Func",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "Func"}}}
			},
			want1: "TestFunc",
		},
		{
			name: "method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{
					Name: &ast.Ident{Name: "method"},
					Recv: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "*Receiver"}}}},
				}}
			},
			want1: "TestReceiver_method",
		},
		{
			name: "Method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{
					Name: &ast.Ident{Name: "Method"},
					Recv: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "*Receiver"}}}},
				}}
			},
			want1: "TestReceiver_Method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.TestName()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.TestName got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_IsMethod(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 bool
	}{
		{
			name: "is not method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{}}
			},
		},
		{
			name: "is method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Recv: &ast.FieldList{}}}
			},
			want1: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.IsMethod()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.IsMethod got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_ReceiverType(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 string
	}{
		{
			name: "is not method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{}}
			},
		},
		{
			name: "is method",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Recv: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "Receiver"}}}}}}
			},
			want1: "Receiver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.ReceiverType()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.ReceiverType got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_ReturnsError(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 bool
	}{
		{
			name: "no results",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Type: &ast.FuncType{}}}
			},
		},
		{
			name: "no error",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
		},
		{
			name: "returns error",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "error"}},
							},
						}},
					},
				}
			},
			want1: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.ReturnsError()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.ReturnsError got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_LastParam(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 *ast.Field
	}{
		{
			name: "no params",
			init: func(*testing.T) *Func {
				return &Func{
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{
							Params: &ast.FieldList{},
						},
					},
				}
			},
		},
		{
			name: "has params",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "s3"}}},
							},
						}},
					},
				}
			},
			want1: &ast.Field{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "s3"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.LastParam()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.LastParam got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_LastResult(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 *ast.Field
	}{
		{
			name: "no results",
			init: func(*testing.T) *Func {
				return &Func{
					Signature: &ast.FuncDecl{Type: &ast.FuncType{}},
				}
			},
		},
		{
			name: "has results",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "s3"}}},
							},
						}},
					},
				}
			},
			want1: &ast.Field{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "s3"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.LastResult()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.LastResult got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestFunc_IsVariadic(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after method run

		want1 bool
	}{
		{
			name: "no params",
			init: func(*testing.T) *Func {
				return &Func{
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{}},
					},
				}
			},
		},
		{
			name: "is not variadic",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ident{Name: "int"}, Names: []*ast.Ident{{Name: "i1"}}},
							},
						}},
					},
				}
			},
		},
		{
			name: "is variadic",
			init: func(*testing.T) *Func {
				return &Func{
					fs: token.NewFileSet(),
					Signature: &ast.FuncDecl{
						Type: &ast.FuncType{Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}, Names: []*ast.Ident{{Name: "s1"}, {Name: "s2"}}},
								{Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "int"}}, Names: []*ast.Ident{{Name: "i3"}}},
							},
						}},
					},
				}
			},
			want1: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receiver := tt.init(t)
			got1 := receiver.IsVariadic()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.IsVariadic got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

func TestFunc_Name(t *testing.T) {
	tests := []struct {
		name string

		init    func(t *testing.T) *Func
		inspect func(r *Func, t *testing.T) //inspects receiver after test run

		want1 string
	}{
		{
			name: "success",
			init: func(*testing.T) *Func {
				return &Func{Signature: &ast.FuncDecl{Name: &ast.Ident{Name: "test"}}}
			},
			want1: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)

			got1 := receiver.Name()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Func.Name got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
