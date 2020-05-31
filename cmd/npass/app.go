package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const envDBKey = "NPASS_DB"

type app struct {
	w  io.Writer
	st store
}

func newApp(ctx context.Context) (*app, error) {
	a := &app{w: os.Stdout}

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

	return a, nil
}

func (a *app) Close() error {
	return a.st.Close()
}

func (a *app) run(ctx context.Context, args []string) error {
	return runMap{
		"new":  runFunc(a.cmdNew),
		"show": runFunc(a.cmdShow),
	}.run(ctx, args)
}

func (a *app) cmdNew(ctx context.Context, args []string) error {
	return nil
}

func (a *app) cmdShow(ctx context.Context, args []string) error {
	return nil
}
