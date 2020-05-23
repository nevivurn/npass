package main

import "github.com/urfave/cli/v2"

func (app *app) cmdPass() *cli.Command {
	cmd := &cli.Command{
		Name: "pass",
		Subcommands: []*cli.Command{
			{Name: "list", Action: app.cmdPassList},
			{Name: "add", Action: app.cmdPassAdd},
			{Name: "del", Action: app.cmdPassDel},
			{Name: "get", Action: app.cmdPassGet},
		},
	}

	return cmd
}

func (app *app) cmdPassList(ctx *cli.Context) error { return nil }
func (app *app) cmdPassAdd(ctx *cli.Context) error  { return nil }
func (app *app) cmdPassDel(ctx *cli.Context) error  { return nil }
func (app *app) cmdPassGet(ctx *cli.Context) error  { return nil }
