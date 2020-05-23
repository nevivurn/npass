package main

import (
	"github.com/urfave/cli/v2"
)

func (app *app) cmdKey() *cli.Command {
	return &cli.Command{
		Name: "key",
		Subcommands: []*cli.Command{
			{Name: "list", Action: app.cmdKeyList},
			{Name: "add", Action: app.cmdKeyAdd},
			{Name: "del", Action: app.cmdKeyDel},
			{Name: "get", Action: app.cmdKeyGet},
			{Name: "pass", Action: app.cmdKeyPass},
		},
	}
}

func (app *app) cmdKeyList(ctx *cli.Context) error { return nil }
func (app *app) cmdKeyAdd(ctx *cli.Context) error  { return nil }
func (app *app) cmdKeyDel(ctx *cli.Context) error  { return nil }
func (app *app) cmdKeyGet(ctx *cli.Context) error  { return nil }
func (app *app) cmdKeyPass(ctx *cli.Context) error { return nil }
