package main

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := newApp()
	app.cliApp.Run(os.Args)
}

type app struct {
	cliApp *cli.App

	store *store
}

func newApp() *app {
	app := &app{}

	app.cliApp = &cli.App{
		Before: app.openDB,

		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:        "db",
				Value:       "",
				DefaultText: "~/.npass.db",
			},
		},

		Commands: []*cli.Command{
			app.cmdKey(),
			app.cmdPass(),
		},
	}

	return app
}

func (app *app) openDB(ctx *cli.Context) error {
	db := ctx.Path("db")
	if db == "" {
		db = filepath.Join(os.Getenv("HOME"), "npass.db")
	}

	st, err := newStore(db)
	if err != nil {
		return err
	}
	app.store = st

	return nil
}
