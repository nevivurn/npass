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

	want := `test-1 CM4ICfq6RLZ/P6qqKElNe5Pr+pk+v1PKJrbTzsbvSHk:
  test-1: [pass]
test-2 LzjhStmiT786jQslhaHcREWoy9vwGOvDqfXHVTfZfxY:
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

	want := "test-1 CM4ICfq6RLZ/P6qqKElNe5Pr+pk+v1PKJrbTzsbvSHk:\n  test-1: [pass]\n"
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

func TestCmdNameKeyFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-none:none"})
	if want := fmt.Errorf("non-existent key %q", "test-none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (name) err = %v; want %v", err, want)
	}
}
func TestCmdNameNameFail(t *testing.T) {
	ctx := context.Background()
	app, _ := testNewApp(t, &testPinentry{})

	err := app.run(ctx, []string{"show", "test-1:none"})
	if want := fmt.Errorf("non-existent pass %q", "test-1:none"); !reflect.DeepEqual(err, want) {
		t.Fatalf("show (name) err = %v; want %v", err, want)
	}
}
