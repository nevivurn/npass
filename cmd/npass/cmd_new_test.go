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
		t.Errorf("new (key) err = %v; want %v", err, errUsage)
	}
}
