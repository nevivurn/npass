package main

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewApp(t *testing.T) {
	oldEnv := os.Getenv(envDBKey)
	os.Setenv(envDBKey, ":memory:")
	defer func() { os.Setenv(envDBKey, oldEnv) }()

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

	if a.w != tmpOut {
		t.Errorf("a.w = %#v; want %#v", a.w, os.Stdout)
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
