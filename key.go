package main

import (
	"github.com/urfave/cli/v2"
)

func cmdKey() *cli.Command {
	return &cli.Command{
		Name: "key",
		Subcommands: []*cli.Command{
			cmdKeyAdd(),
		},
	}
}

func cmdKeyAdd() *cli.Command {
	cmd := &cli.Command{
		Name: "add",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Required: true,
			},
		},
	}

	cmd.Action = func(ctx *cli.Context) error {
		return nil
	}

	return cmd
}
