package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestCmdShow(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "", ""})
	if !errors.Is(err, errUsage) {
		t.Errorf("show err = %v; want %v", err, errUsage)
	}

	err = app.run(ctx, []string{"show", "INVALID"})
	if !errors.Is(err, errIdentifier) {
		t.Errorf("show err = %v; want %v", err, errIdentifier)
	}
}

func TestCmdShowAll(t *testing.T) {
	ctx := context.Background()
	app, out := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `test-1 5M60s3mDoFQJQOkZxyn0nOo2VuKzdUJZU60j3Ymkny4:
  test-1: [pass]
test-2 VIFZnL4uDjJIwU61yIN0NV5ORYdehWU2PYeTwMwDvwc:
  test-2: [pass]
`
	if out.String() != want {
		t.Errorf("show (all) out = %q; want %q", out.String(), want)
	}
}

func TestCmdShowKey(t *testing.T) {
	ctx := context.Background()
	app, out := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "test-1 5M60s3mDoFQJQOkZxyn0nOo2VuKzdUJZU60j3Ymkny4:\n  test-1: [pass]\n"
	if out.String() != want {
		t.Errorf("show (key) out = %q; want %q", out.String(), want)
	}
}

func TestCmdShowKeyKeyFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-none"})
	if want := fmt.Errorf("non-existent key %q", "test-none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (key) err = %v; want %v", err, want)
	}
}

func TestCmdShowName(t *testing.T) {
	ctx := context.Background()
	app, out := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1:test-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "test-1:test-1:pass\n"
	if out.String() != want {
		t.Errorf("show (name) out = %q; want %q", out.String(), want)
	}
}

func TestCmdShowNameKeyFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-none:none"})
	if want := fmt.Errorf("non-existent key %q", "test-none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (name) err = %v; want %v", err, want)
	}
}
func TestCmdShowNameNameFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1:none"})
	if want := fmt.Errorf("non-existent pass %q", "test-1:none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (name) err = %v; want %v", err, want)
	}
}

func TestCmdShowPass(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{pass: "pass-1"}
	app, out := testNewApp(t, pin)

	err := app.run(ctx, []string{"show", "test-1:test-1:pass"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "pass-1\n"; out.String() != want {
		t.Errorf("show (pass) out = %q; want %q", out.String(), want)
	}
}

func TestCmdShowPassKeyFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-none:none:none"})
	if want := fmt.Errorf("non-existent key %q", "test-none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (pass) err = %v; want %v", err, want)
	}
}
func TestCmdShowPassNameFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1:none:none"})
	if want := fmt.Errorf("non-existent pass %q", "test-1:none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (pass) err = %v; want %v", err, want)
	}
}

func TestCmdShowPassTypInvalidFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1:test-1:none"})
	if !errors.Is(err, errInvalidPassType) {
		t.Fatalf("show (pass) err = %v; want %v", err, errInvalidPassType)
	}
}
func TestCmdShowPassPassFail(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{pass: "incorrect"}
	app, _ := testNewApp(t, pin)

	err := app.run(ctx, []string{"show", "test-1:test-1:pass"})
	if !errors.Is(err, errTestPinentryVerify) {
		t.Fatalf("show (pass) err = %v; want %v", err, errTestPinentryVerify)
	}
}

func TestCmdShowPassPinFail(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{err: errors.New("testing error")}
	app, _ := testNewApp(t, pin)

	err := app.run(ctx, []string{"show", "test-1:test-1:pass"})
	if !errors.Is(err, pin.err) {
		t.Fatalf("show (pass) err = %v; want %v", err, pin.err)
	}
}
