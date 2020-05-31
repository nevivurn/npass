package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	ctx := context.Background()
	ctx, cancel := withShutdown(ctx)
	defer cancel()

	a, err := newApp(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
	defer a.Close()

	if err := a.run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func withShutdown(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	die := make(chan os.Signal, 1)
	signal.Notify(die, os.Interrupt)

	go func() {
		select {
		case <-die:
		case <-ctx.Done():
		}
		cancel()
		signal.Stop(die)
	}()

	return ctx, cancel
}
