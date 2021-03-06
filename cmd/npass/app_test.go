package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/nevivurn/npass/pkg/pinentry"
)

func TestNewApp(t *testing.T) {
	oldEnv := os.Getenv(envDBKey)
	os.Setenv(envDBKey, ":memory:")
	defer func() { os.Setenv(envDBKey, oldEnv) }()

	oldStdin := os.Stdin
	discard, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0777)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer discard.Close()
	os.Stdin = discard
	defer func() { os.Stdin = oldStdin }()

	// Capture output for initialization message
	oldStdout := os.Stdout
	tmpOut, err := ioutil.TempFile("", "npass-test-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(tmpOut.Name())
	defer tmpOut.Close()
	os.Stdout = tmpOut
	defer func() { os.Stdout = oldStdout }()

	a, err := newApp(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Close()

	if a.r != discard {
		t.Errorf("a.r = %#v; want %#v", a.r, discard)
	}

	if a.w != tmpOut {
		t.Errorf("a.w = %#v; want %#v", a.w, tmpOut)
	}

	_, err = tmpOut.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := ioutil.ReadAll(tmpOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "Initialized new db at :memory:\n"; string(out) != want {
		t.Errorf("output = %q; want %q", string(out), want)
	}

	ok, err := a.st.checkSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("checkSchema() = %t; want %t", ok, true)
	}

	if a.pin != pinentry.External {
		t.Errorf("a.pin = %#v; want %#v", a.pin, pinentry.External)
	}
}

func TestNewAppDefault(t *testing.T) {
	// Just in case the env var is set during tests
	oldEnv := os.Getenv(envDBKey)
	os.Unsetenv(envDBKey)
	defer func() { os.Setenv(envDBKey, oldEnv) }()

	// Silence message about initializing db
	oldStdout := os.Stdout
	discard, err := os.OpenFile(os.DevNull, os.O_RDWR, 0777)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer discard.Close()
	os.Stdout = discard
	defer func() { os.Stdout = oldStdout }()

	// Avoid actually creating db files
	defaultStoreArgs["mode"] = "memory"
	defer delete(defaultStoreArgs, "mode")

	a, err := newApp(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Close()

	st, err := newStore(filepath.Join(os.Getenv("HOME"), ".npass.db"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer st.Close()

	ok, err := st.checkSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("checkSchema() = %t; want %t", ok, true)
	}
}

type testPinentry struct {
	confirm bool
	pass    string
	err     error
}

var errTestPinentryVerify = errors.New("pinentry: verification error (testing)")

func (tp testPinentry) Confirm(context.Context, string) (bool, error) {
	return tp.confirm, tp.err
}
func (tp testPinentry) NewPass(context.Context, string) (string, error) {
	return tp.pass, tp.err
}
func (tp testPinentry) AskPass(_ context.Context, _ string, f func(string) bool) (string, error) {
	if tp.err != nil {
		return "", tp.err
	}
	if !f(tp.pass) {
		return "", errTestPinentryVerify
	}
	return tp.pass, nil
}

func testNewApp(t *testing.T, pin pinentry.Pinentry) (*app, *bytes.Buffer) {
	fastKDF = true
	t.Cleanup(func() { fastKDF = false })

	buf := &bytes.Buffer{}

	schema, err := ioutil.ReadFile("testdata/test_schema.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	st := testStore(t)
	if _, err := st.Exec(string(schema)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return &app{
		w:   buf,
		st:  st,
		pin: pin,
	}, buf
}
