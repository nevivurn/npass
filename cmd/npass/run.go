package main

import (
	"context"
	"errors"
)

var errUsage = errors.New("incorrect usage")

type runner interface {
	run(context.Context, []string) error
}

type runFunc func(context.Context, []string) error

var _ runner = runFunc(nil) // Static interface check

func (rf runFunc) run(ctx context.Context, args []string) error {
	return rf(ctx, args)
}

type runMap map[string]runner

var _ runner = runMap(nil) // Static interface check

func (rm runMap) run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errUsage
	}

	run, ok := rm[args[0]]
	if !ok {
		return errUsage
	}

	return run.run(ctx, args[1:])
}
