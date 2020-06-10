package main

import "context"

func (a *app) cmdShow(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return errUsage
	}

	var (
		key, name, typ string
		err            error
	)
	if len(args) == 1 {
		key, name, typ, err = parseIdentifier(args[0])
		if err != nil {
			return err
		}
	}

	if key == "" && name == "" && typ == "" {
		err = a.cmdShowAll(ctx)
	} else if key != "" && name == "" && typ == "" {
		err = a.cmdShowKey(ctx, key)
	} else if key != "" && name != "" && typ == "" {
		err = a.cmdShowName(ctx, key, name)
	} else if key != "" && name != "" && typ != "" {
		err = a.cmdShowPass(ctx, key, name, typ)
	}

	return err
}

func (a *app) cmdShowAll(ctx context.Context) error {
	panic("not yet implemented")
}

func (a *app) cmdShowKey(ctx context.Context, key string) error {
	panic("not yet implemented")
}

func (a *app) cmdShowName(ctx context.Context, key, name string) error {
	panic("not yet implemented")
}

func (a *app) cmdShowPass(ctx context.Context, key, name, typ string) error {
	panic("not yet implemented")
}
