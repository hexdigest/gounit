package main

import (
	"reflect"
	"testing"
)

func TestLinesNumbers_Set(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *LinesNumbers
		inspect func(r *LinesNumbers, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "invalid values",
			init: func(*testing.T) *LinesNumbers {
				return &LinesNumbers{}
			},
			args: func(*testing.T) args {
				return args{value: "a,b"}
			},
			wantErr: true,
		},
		{
			name: "valid values",
			init: func(*testing.T) *LinesNumbers {
				return &LinesNumbers{}
			},
			args: func(*testing.T) args {
				return args{value: "1,2"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.Set(tArgs.value)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("LinesNumbers.Set error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestLinesNumbers_String(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *LinesNumbers
		inspect func(r *LinesNumbers, t *testing.T) //inspects receiver after test run

		want1 string
	}{
		{
			name: "ok",
			init: func(t *testing.T) *LinesNumbers {
				return &LinesNumbers{1, 2}
			},
			want1: "[1 2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.String()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("LinesNumbers.String got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestFunctionsList_Set(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *FunctionsList
		inspect func(r *FunctionsList, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T) //use for more precise error evaluation after test
	}{
		{
			name: "invalid values",
			init: func(t *testing.T) *FunctionsList {
				return &FunctionsList{}
			},
			args: func(t *testing.T) args {
				return args{value: "!!!"}
			},
			wantErr: true,
		},
		{
			name: "valid values",
			init: func(t *testing.T) *FunctionsList {
				return &FunctionsList{}
			},
			args: func(t *testing.T) args {
				return args{value: "func1,func2"}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			err := receiver.Set(tArgs.value)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("FunctionsList.Set error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestFunctionsList_String(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *FunctionsList
		inspect func(r *FunctionsList, t *testing.T) //inspects receiver after test run

		want1 string
	}{
		{
			name: "ok",
			init: func(t *testing.T) *FunctionsList {
				return &FunctionsList{"func1", "func2"}
			},
			want1: "[func1 func2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.String()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("FunctionsList.String got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
