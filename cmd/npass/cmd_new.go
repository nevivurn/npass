package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var cmdNew = &cli.Command{
	Name:    "new",
	Aliases: []string{"n"},

	Usage:     "Creates a new key or password",
	ArgsUsage: "KEY[:NAME:TYPE]",

	Action: cmdNewAction,
}

func cmdNewAction(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		return cli.Exit("new accepts exactly 1 argument", 1)
	}

	p, err := parsePass(ctx.Args().First())
	if err != nil {
		return cli.Exit(err, 1)
	}

	if p.name != "" && p.typ == "" {
		return cli.Exit("type must be specified when creating pass", 1)
	}

	if p.typ == "" {
		err = cmdNewCreateKey(ctx, p)
	} else {
		err = cmdNewCreatePass(ctx, p)
	}
	return err
}

func cmdNewCreateKey(ctx *cli.Context, p passName) error {
	fmt.Fprintln(ctx.App.Writer, "create key")
	return nil
}

func cmdNewCreatePass(ctx *cli.Context, p passName) error {
	fmt.Fprintln(ctx.App.Writer, "create pass")
	return nil
}
