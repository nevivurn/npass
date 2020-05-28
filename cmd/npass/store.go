package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/nevivurn/npass/store"
)

type ctxKey string

const ctxStore ctxKey = "store"

func flagStoreBefore(ctx *cli.Context) error {
	dbName := ctx.Path("db")
	if !ctx.IsSet("db") {
		dbName = filepath.Join(os.Getenv("HOME"), ".npass.db")
	}

	if _, err := os.Stat(dbName); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(ctx.App.ErrWriter, "initializing new DB at", dbName)
	} else if err != nil {
		return err
	}

	st, err := store.New(dbName)
	if err != nil {
		return err
	}
	ctx.Context = context.WithValue(ctx.Context, ctxStore, st)

	return nil
}

func flagStoreAfter(ctx *cli.Context) error {
	st, ok := ctx.Context.Value(ctxStore).(*store.Store)
	if ok {
		return st.Close()
	}
	return nil
}
