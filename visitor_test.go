package gounit

import (
	"go/ast"
	"reflect"
	"testing"
)

func Test_NewVisitor(t *testing.T) {
	type args struct {
		match matchFunc
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		got1 *Visitor
	}{
		{
			name: "success",
			args: func(*testing.T) args { return args{} },
			got1: &Visitor{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1 := NewVisitor(tArgs.match)

			if !reflect.DeepEqual(got1, tt.got1) {
				t.Errorf("NewVisitor got1 = %v, got1: %v", got1, tt.got1)
			}

		})
	}
}

func Test_Visitor_Visit(t *testing.T) {
	type args struct {
		node ast.Node
	}

	var notFoundVisitor = NewVisitor(func(*ast.FuncDecl) bool {
		return false
	})

	tests := []struct {
		name    string
		args    func(t *testing.T) args
		init    func(t *testing.T) *Visitor
		inspect func(r *Visitor, t *testing.T) //inspects receiver after method run

		got1 ast.Visitor
	}{
		{
			name: "func not found",
			init: func(*testing.T) *Visitor { return notFoundVisitor },
			args: func(*testing.T) args { return args{} },
			got1: notFoundVisitor,
		},
		{
			name: "func found",
			init: func(*testing.T) *Visitor {
				return &Visitor{match: func(*ast.FuncDecl) bool { return true }}
			},
			args: func(*testing.T) args {
				return args{
					node: &ast.FuncDecl{},
				}
			},
			inspect: func(v *Visitor, t *testing.T) {
				if v.found == nil {
					t.Errorf("expected non-nil v.found")
				}
			},
			got1: nil,
		},
		{
			name: "func was already found",
			init: func(*testing.T) *Visitor {
				return &Visitor{found: &ast.FuncDecl{}}
			},
			args: func(*testing.T) args {
				return args{
					node: nil,
				}
			},
			got1: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			receiver := tt.init(t)
			got1 := receiver.Visit(tArgs.node)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.got1) {
				t.Errorf("Visitor.Visit got1 = %v, got1: %v", got1, tt.got1)
			}

		})
	}
}
