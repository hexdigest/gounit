package gounit

import (
	"go/ast"
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
