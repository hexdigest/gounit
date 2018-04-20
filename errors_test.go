package gounit

import (
	"reflect"
	"testing"
)

func TestError_Format(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) GenericError
		inspect func(r GenericError, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "success",
			init: func(*testing.T) GenericError {
				return GenericError("%d - %s = 0")
			},
			args: func(*testing.T) args {
				return args{
					args: []interface{}{1, "1"},
				}
			},
			inspectErr: func(err error, t *testing.T) {
				if err.Error() != "1 - 1 = 0" {
					t.Errorf("unexpected result: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.Format(tArgs.args...)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	type args struct {
		code   int
		format string
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 GenericError
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{
					code:   1,
					format: "format",
				}
			},
			want1: GenericError("format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := GenericError(tArgs.format)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewError got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
