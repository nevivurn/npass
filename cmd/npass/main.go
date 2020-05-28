package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"
)

var (
	version string
	cliApp  = &cli.App{
		Name: "npass",

		Usage:   "A multi-device password manager",
		Version: version,

		Writer:    os.Stdout,
		ErrWriter: os.Stderr,

		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "db",
				Aliases: []string{"d"},

				Usage: "path to the npass database",

				EnvVars:     []string{"NPASS_DB"},
				DefaultText: "~/.npass.db",
			},
		},
		// Open DB and set ctxStore, close on exit.
		Before: flagStoreBefore,
		After:  flagStoreAfter,

		Commands: []*cli.Command{
			cmdNew,
			//cmdShow,
			//cmdRm,
		},
	}
)

func main() {
	ctx := context.Background()
	ctx, cancel := withInterrupt(ctx)
	defer cancel()

	err := cliApp.RunContext(ctx, os.Args)
	if err != nil {
		cancel()
		os.Exit(1)
	}
}

func withInterrupt(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	die := make(chan os.Signal, 1)
	signal.Notify(die, os.Interrupt)

	go func() {
		select {
		case <-ctx.Done():
		case <-die:
		}
		cancel()
		signal.Stop(die)
	}()

	return ctx, cancel
}
