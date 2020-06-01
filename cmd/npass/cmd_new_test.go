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

func TestCmdNewKey(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{pass: "pass"}
	app, out := testNewApp(t, pin)

	err := app.run(ctx, []string{"new", "key"})
	if !errors.Is(err, pin.err) {
		t.Errorf("new (key) err = %v; want %v", err, errUsage)
	}

	var key, pub, priv string
	err = app.st.QueryRow(`SELECT name, public, private FROM keys`).Scan(&key, &pub, &priv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key != "key" {
		t.Errorf("inserted name = %q; want %q", key, "key")
	}
	if want := fmt.Sprintf("created new key %q: %s\n", key, pub); out.String() != want {
		t.Errorf("out = %q; want %q", out.String(), want)
	}
}

func TestCmdNewKeyDuplicateFail(t *testing.T) {
	ctx := context.Background()
	pin := &testPinentry{pass: "pass"}
	app, _ := testNewApp(t, pin)

	err := app.run(ctx, []string{"new", "key"})
	if !errors.Is(err, pin.err) {
		t.Errorf("new (key) err = %v; want %v", err, errUsage)
	}

	err = app.run(ctx, []string{"new", "key"})
	if want := fmt.Errorf("duplicate key %q", "key"); !reflect.DeepEqual(err, want) {
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
