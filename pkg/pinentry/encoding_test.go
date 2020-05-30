package pinentry

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestPercentEncode(t *testing.T) {
	tests := map[string]string{
		"hello, world!": "hello, world!",
		"%\r\n":         "%25%0D%0A",
	}

	for tc, want := range tests {
		got := percentEncode(tc)
		if got != want {
			t.Errorf("percentEncode(%q) = %q; want %q", tc, got, want)
		}
	}
}

func TestPercentDecode(t *testing.T) {
	type testCase struct {
		s   string
		err error
	}
	tests := map[string]testCase{
		"hello, world!": {"hello, world!", nil},
		"%25%0D%0A":     {"%\r\n", nil},
		"%":             {"", io.EOF},
		"%zz":           {"", fmt.Errorf("invalid percent-encoding")},
	}

	for tc, want := range tests {
		got, err := percentDecode(tc)
		if out := (testCase{got, err}); !reflect.DeepEqual(out, want) {
			t.Errorf("percentDecode(%q) = %#v; want %#v", tc, out, want)
		}
	}
}

func TestSend(t *testing.T) {
	var out bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&bytes.Reader{}), bufio.NewWriter(&out))

	err := send(rw, "COMMAND", "hello", "%", "\r", "\n")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	want := "COMMAND hello %25 %0D %0A\n"
	if out.String() != want {
		t.Errorf("send() = %q; want %q", out.String(), want)
	}
}

func TestRecv(t *testing.T) {
	type testCase struct {
		s   string
		err error
	}
	tests := map[string]testCase{
		"OK Pleased to meet you\n":              {},
		"OK\n":                                  {},
		"ERR 0 Hello there\n":                   {"", fmt.Errorf("pinentry error: 0 Hello there")},
		"S Ignore me\n# Ignore me too\nOK ok\n": {},
		"D hello %25 %0D %0A\nOK ok\n":          {"hello % \r \n", nil},
		"":                                      {"", io.EOF},
		"\n":                                    {"", fmt.Errorf("invalid data from pinentry")},
		"D %OK\nOK\n": {"", fmt.Errorf("invalid data from pinentry: %w",
			fmt.Errorf("invalid percent-encoding"))},
	}

	for tc, want := range tests {
		rd := strings.NewReader(tc)
		rw := bufio.NewReadWriter(bufio.NewReader(rd), nil)

		got, err := recv(rw)
		if out := (testCase{got, err}); !reflect.DeepEqual(out, want) {
			fmt.Println(err, want.err)
			t.Errorf("recv(%q) = %#v; want %#v", tc, out, want)
		}
	}
}
