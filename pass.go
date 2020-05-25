package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/nevivurn/npass/pinentry"
	"github.com/nevivurn/npass/store"
)

const (
	flagPassKey  = "key"
	flagPassName = "name"
)

func getDecryptedKey(ctx context.Context, key *store.Key) (*[32]byte, error) {
	if len(key.Private) < 16 {
		return nil, fmt.Errorf("invalid key data in database")
	}
	salt, privEnc := key.Private[:16], key.Private[16:]

	var priv []byte
	_, err := pinentry.ReadPassword(
		ctx,
		pinentry.Prompt(fmt.Sprintf("Enter password for key %q", key.Name)),
		pinentry.Verify(func(pass string) bool {
			var keyArr [32]byte
			copy(keyArr[:], passKey([]byte(pass), salt))

			var ok bool
			priv, ok = secretbox.Open(nil, privEnc, &[24]byte{}, &keyArr)
			return ok
		}),
	)
	if err != nil {
		return nil, err
	}

	var privArr [32]byte
	if len(priv) != len(privArr) {
		return nil, fmt.Errorf("invalid key data in database")
	}
	copy(privArr[:], priv)

	return &privArr, nil
}

func cmdPass() *cli.Command {
	cmd := &cli.Command{
		Name: "pass",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPassKey,
				Value: "default",
			},
		},

		Subcommands: []*cli.Command{
			cmdPassPut(),
		},
	}

	return cmd
}

func cmdPassPut() *cli.Command {
	cmd := &cli.Command{
		Name: "put",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     flagPassName,
				Required: true,
			},
		},
	}

	cmd.Action = func(ctx *cli.Context) error {
		st, err := store.New(ctx.Path(flagDB))
		if err != nil {
			return err
		}
		defer st.Close()

		key, err := st.KeyFindName(ctx.Context, ctx.String(flagPassKey))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("key %q does not exist", ctx.String(flagPassKey))
		}
		if err != nil {
			return err
		}

		name := ctx.String(flagPassName)

		_, err = st.PassFindKeyName(ctx.Context, key.ID, name)
		if err == nil {
			return fmt.Errorf("pass %q already exist for key %q", name, key.Name)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		pass, err := pinentry.ReadPassword(
			ctx.Context,
			pinentry.Prompt(fmt.Sprintf("Enter password for %q", name)),
			pinentry.Confirm(),
		)
		if err != nil {
			return err
		}

		var pubArr [32]byte
		if len(key.Public) != len(pubArr) {
			return fmt.Errorf("invalid key data in database")
		}
		copy(pubArr[:], key.Public)

		passEnc, err := box.SealAnonymous(nil, []byte(pass), &pubArr, rand.Reader)
		if err != nil {
			return err
		}

		st.PassPut(ctx.Context, &store.Pass{
			KeyID: key.ID,
			Name:  name,
			Data:  passEnc,
		})

		return nil
	}

	return cmd
}
