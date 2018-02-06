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
		init    func(t *testing.T) Error
		inspect func(r Error, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "success",
			init: func(*testing.T) Error {
				return Error{format: "%d - %s = 0", code: 3}
			},
			args: func(*testing.T) args {
				return args{
					args: []interface{}{1, "1"},
				}
			},
			wantErr: true,
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

			if (err != nil) != tt.wantErr {
				t.Fatalf("Error.Format error = %v, wantErr: %t", err, tt.wantErr)
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

		want1 Error
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{
					code:   1,
					format: "format",
				}
			},
			want1: Error{code: 1, format: "format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := NewError(tArgs.code, tArgs.format)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewError got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestError_Code(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) Error
		inspect func(r Error, t *testing.T) //inspects receiver after test run

		want1 int
	}{
		{
			name: "success",
			init: func(*testing.T) Error {
				return Error{code: 5}
			},
			want1: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.Code()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Error.Code got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
