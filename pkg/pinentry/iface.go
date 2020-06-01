package pinentry

import "context"

// Pinentry is the interface
type Pinentry interface {
	Confirm(context.Context, string) (bool, error)
	NewPass(context.Context, string) (string, error)
	AskPass(context.Context, string, func(string) bool) (string, error)
}

type pinentry struct{}

// External is the external (default) pinentry implementation.
var External Pinentry = pinentry{}

func (pinentry) Confirm(ctx context.Context, prompt string) (bool, error) {
	return Confirm(ctx, prompt)
}

func (pinentry) NewPass(ctx context.Context, prompt string) (string, error) {
	return NewPass(ctx, prompt)
}

func (pinentry) AskPass(ctx context.Context, prompt string, verify func(string) bool) (string, error) {
	return AskPass(ctx, prompt, verify)
}
