package main

import (
	"context"
	"errors"
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
