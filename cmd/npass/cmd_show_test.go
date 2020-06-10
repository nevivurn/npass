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
