package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			cmdKey(),
		},
	}
}
