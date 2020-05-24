package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nevivurn/npass/pinentry"
	"github.com/urfave/cli/v2"
)

func main() {
	pass, err := pinentry.ReadPassword(context.Background())
	fmt.Printf("%q %v\n", pass, err)

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
