package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestCmdNew(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"new"})
	if !errors.Is(err, errUsage) {
		t.Errorf("new err = %v; want %v", err, errUsage)
	}

	err = app.run(ctx, []string{"new", "", ""})
	if !errors.Is(err, errUsage) {
		t.Errorf("new err = %v; want %v", err, errUsage)
	}

	err = app.run(ctx, []string{"new", "key:name"})
	if !errors.Is(err, errUsage) {
		t.Errorf("new err = %v; want %v", err, errUsage)
	}

	err = app.run(ctx, []string{"new", "INVALID"})
	if !errors.Is(err, errIdentifier) {
		t.Errorf("new err = %v; want %v", err, errIdentifier)
	}
}

func TestCmdNewKeyDuplicateFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"new", "test-1"})
	if want := fmt.Errorf("duplicate key %q", "test-1"); !reflect.DeepEqual(err, want) {
		t.Errorf("new (key) err = %v; want %v", err, want)
	}
}

func TestCmdNewKeyPinFail(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{err: errors.New("testing error")}
	app, _ := testNewApp(t, pin)

	err := app.run(ctx, []string{"new", "key"})
	if !errors.Is(err, pin.err) {
		t.Errorf("new (key) err = %v; want %v", err, pin.err)
	}
}

func TestCmdNewPassKeyFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"new", "test-none:name:pass"})
	if want := fmt.Errorf("non-existent key %q", "test-none"); !reflect.DeepEqual(err, want) {
		t.Errorf("new (pass) err = %v; want %v", err, want)
	}
}

func TestCmdNewPassTypeFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"new", "test-1:name:none"})
	if want := errInvalidPassType; !reflect.DeepEqual(err, want) {
		t.Errorf("new (pass) err = %v; want %v", err, want)
	}
}

func TestCmdNewPassDuplicateFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"new", "test-1:test-1:pass"})
	if want := fmt.Errorf("duplicate pass %q", "test-1:test-1:pass"); !reflect.DeepEqual(err, want) {
		t.Errorf("new (pass) err = %v; want %v", err, want)
	}
}

func TestCmdNewPassPinFail(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{err: errors.New("testing error")}
	app, _ := testNewApp(t, pin)

	err := app.run(ctx, []string{"new", "test-1:name:pass"})
	if !errors.Is(err, pin.err) {
		t.Errorf("new (pass) err = %v; want %v", err, pin.err)
	}
}
