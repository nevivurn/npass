package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nevivurn/npass/store"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	cmd := &cli.App{
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:        "db",
				EnvVars:     []string{"NPASS_DB"},
				DefaultText: "~/.npass.db",
			},
		},

		Commands: []*cli.Command{
			cmdInit(),
		},
	}

	cmd.Before = func(ctx *cli.Context) error {
		if !ctx.IsSet("db") {
			return ctx.Set("db", filepath.Join(os.Getenv("HOME"), ".npass.db"))
		}
		return nil
	}

	return cmd
}

func cmdInit() *cli.Command {
	cmd := &cli.Command{
		Name: "init",
	}

	cmd.Action = func(ctx *cli.Context) error {
		name := ctx.Path("db")

		_, err := os.Stat(name)
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("database already exists")
		}

		st, err := store.NewInit(name)
		if err != nil {
			return err
		}
		defer st.Close()

		fmt.Fprintln(ctx.App.Writer, "initialized a new db at", name)
		return nil
	}

	return cmd
}
