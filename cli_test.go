package gounit

import (
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	type args struct {
		r io.Reader
		w io.Writer
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 *CLI
	}{
		{
			name: "success",
			args: func(*testing.T) args {
				return args{}
			},
			want1: &CLI{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := NewCLI(tArgs.r, tArgs.w)

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewCLI got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestCLI_Read(t *testing.T) {

	cli := NewCLI(
		strings.NewReader("world\n"),
		newExpectPrefixWriter(t, "hello:\n"),
	)

	var s string
	err := cli.Read("hello", &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if s != "world" {
		t.Errorf("unexpected value: %s", s)
	}
}

func TestCLI_ReadUntil(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cli := NewCLI(
			strings.NewReader("successeof"),
			newExpectPrefixWriter(t, "value followed by eof:\n"),
		)

		b, err := cli.ReadUntil("value", "eof")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if string(b) != "success" {
			t.Errorf("unexpected result: %s", string(b))
		}
	})

	t.Run("scan failed", func(t *testing.T) {
		cli := NewCLI(
			errorReader{errors.New("read error")},
			ioutil.Discard,
		)

		_, err := cli.ReadUntil("value", "eof")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func Test_splitByDelimiter(t *testing.T) {
	f := splitByDelimiter("delim")

	t.Run("at eof and no data", func(t *testing.T) {
		advance, token, err := f([]byte{}, true)
		if advance != 0 {
			t.Errorf("got advance: %d, want: 0", advance)
		}
		if token != nil {
			t.Errorf("got token: %s, want: nil", string(token))
		}
		if err != nil {
			t.Errorf("got error: %v, want: nil", err)
		}
	})

	t.Run("found delimiter", func(t *testing.T) {
		advance, token, err := f([]byte("123delim345"), false)
		if advance != 8 {
			t.Errorf("got advance: %d, want: 8", advance)
		}
		if string(token) != "123" {
			t.Errorf("got token: %s, want: 123", string(token))
		}
		if err != nil {
			t.Errorf("got error: %v, want: nil", err)
		}
	})

	t.Run("at eof", func(t *testing.T) {
		advance, token, err := f([]byte("123"), true)
		if advance != 3 {
			t.Errorf("got advance: %d, want: 3", advance)
		}
		if string(token) != "123" {
			t.Errorf("got token: %s, want: 123", string(token))
		}
		if err != nil {
			t.Errorf("got error: %v, want: nil", err)
		}
	})

	t.Run("at eof", func(t *testing.T) {
		advance, token, err := f([]byte("123"), false)
		if advance != 0 {
			t.Errorf("got advance: %d, want: 0", advance)
		}
		if token != nil {
			t.Errorf("got token: %s, want: nil", string(token))
		}
		if err != nil {
			t.Errorf("got error: %v, want: nil", err)
		}
	})
}
