package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nevivurn/npass/pkg/pinentry"
)

const envDBKey = "NPASS_DB"

type app struct {
	r   io.Reader
	w   io.Writer
	st  store
	pin pinentry.Pinentry
}

func newApp(ctx context.Context) (*app, error) {
	a := &app{
		r: os.Stdin,
		w: os.Stdout,
	}

	db := os.Getenv(envDBKey)
	if db == "" {
		db = filepath.Join(os.Getenv("HOME"), ".npass.db")
	}

	st, err := newStore(db, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}
	a.st = st

	ok, err := st.checkSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	if !ok {
		err := st.initSchema(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not initialize db: %w", err)
		}
		fmt.Fprintf(a.w, "Initialized new db at %s\n", db)
	}

	a.pin = pinentry.External

	return a, nil
}

func (a *app) Close() error {
	return a.st.Close()
}

func (a *app) run(ctx context.Context, args []string) error {
	return runMap{
		"new": runFunc(a.cmdNew),
	}.run(ctx, args)
}
