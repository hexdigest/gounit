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

func Test_GetOptions(t *testing.T) {
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
					stderr: newExpectPrefixWriter(t, "missing line number or function name"),
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
			name: "success",
			args: func(t *testing.T) args {
				return args{
					stdout:    ioutil.Discard,
					arguments: []string{"-l", "10", "-i", "input.go", "-c", "TODO"},
				}
			},
			want1: Options{
				LineNumber: 10,
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
