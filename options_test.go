package gounit

import (
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func expectExitCode(t *testing.T, expectedCode int) exitFunc {
	return func(code int) {
		if code != expectedCode {
			t.Errorf("got exit code: %d, want: %d", code, expectedCode)
		}

		t.Skip()
	}
}

func TestGetOptions(t *testing.T) {
	type args struct {
		arguments []string
		stdout    io.Writer
		stderr    io.Writer
		exit      exitFunc
	}

	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 Options
	}{
		{
			name: "show help",
			args: func(t *testing.T) args {
				return args{
					arguments: []string{"-h"},
					stdout:    newExpectPrefixWriter(t, "Usage of gounit:"),
					stderr:    ioutil.Discard,
					exit:      expectExitCode(t, 0),
				}
			},
		},
		{
			name: "missing line number",
			args: func(t *testing.T) args {
				return args{
					stderr: newExpectPrefixWriter(t, "missing line numbers or function names"),
					stdout: ioutil.Discard,
					exit:   expectExitCode(t, 2),
				}
			},
		},
		{
			name: "missing input file",
			args: func(t *testing.T) args {
				return args{
					stderr:    newExpectPrefixWriter(t, "missing input file"),
					stdout:    ioutil.Discard,
					exit:      expectExitCode(t, 2),
					arguments: []string{"-l", "10"},
				}
			},
		},
		{
			name: "cli",
			args: func(t *testing.T) args {
				return args{
					stdout:    ioutil.Discard,
					arguments: []string{"-cli"},
				}
			},
			want1: Options{UseCLI: true},
		},
		{
			name: "success",
			args: func(t *testing.T) args {
				return args{
					stdout:    ioutil.Discard,
					arguments: []string{"-l", "10", "-i", "input.go", "-c", "TODO"},
				}
			},
			want1: Options{
				Lines:      LinesNumbers{10},
				InputFile:  "input.go",
				OutputFile: "input_test.go",
				Comment:    "TODO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got1 := GetOptions(tArgs.arguments, tArgs.stdout, tArgs.stderr, tArgs.exit)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetOptions got1 = %v, want1: %v", got1, tt.want1)
			}

		})
	}
}

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

func TestLinesNumbers_Include(t *testing.T) {
	type args struct {
		line int
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) LinesNumbers
		inspect func(r LinesNumbers, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1 bool
	}{
		{
			name: "true",
			init: func(t *testing.T) LinesNumbers {
				return LinesNumbers{1, 2, 3}
			},
			args: func(*testing.T) args {
				return args{line: 2}
			},
			want1: true,
		},
		{
			name: "false",
			init: func(t *testing.T) LinesNumbers {
				return LinesNumbers{1, 2, 3}
			},
			args: func(*testing.T) args {
				return args{line: 7}
			},
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1 := receiver.Include(tArgs.line)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("LinesNumbers.Include got1 = %v, want1: %v", got1, tt.want1)
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

func TestFunctionsList_Include(t *testing.T) {
	type args struct {
		function string
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) FunctionsList
		inspect func(r FunctionsList, t *testing.T) //inspects receiver after test run

		args func(t *testing.T) args

		want1 bool
	}{
		{
			name: "true",
			init: func(t *testing.T) FunctionsList {
				return FunctionsList{"func1", "func2", "func3"}
			},
			args: func(*testing.T) args {
				return args{function: "func2"}
			},
			want1: true,
		},
		{
			name: "false",
			init: func(t *testing.T) FunctionsList {
				return FunctionsList{"func1", "func2", "func3"}
			},
			args: func(*testing.T) args {
				return args{function: "func5"}
			},
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1 := receiver.Include(tArgs.function)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("FunctionsList.Include got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}
