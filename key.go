package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/nevivurn/npass/pinentry"
	"github.com/nevivurn/npass/store"
)

const (
	flagKeyName = "name"
)

func passKey(pass []byte, salt []byte) []byte {
	return argon2.IDKey(pass, salt, 1, 6<<20, 8, 32)
}

func cmdKey() *cli.Command {
	cmd := &cli.Command{
		Name: "key",
		Subcommands: []*cli.Command{
			cmdKeyPut(),
		},
	}

	return cmd
}

func cmdKeyPut() *cli.Command {
	cmd := &cli.Command{
		Name: "put",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     flagKeyName,
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

		name := ctx.String(flagKeyName)

		_, err = st.KeyFindName(ctx.Context, name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil {
			return fmt.Errorf("key %q already exists", name)
		}

		pub, priv, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		pass, err := pinentry.ReadPassword(ctx.Context, fmt.Sprintf("Enter password for key %q", name))
		if err != nil {
			return err
		}
		_, err = pinentry.ReadPasswordVerify(ctx.Context, "Confirm password",
			func(s string) bool {
				return s == pass
			})
		if err != nil {
			return err
		}

		salt := make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return err
		}

		key := passKey([]byte(pass), salt)
		keyArr := [32]byte{}
		copy(keyArr[:], key)

		privEnc := secretbox.Seal(salt, priv[:], &[24]byte{}, &keyArr)

		_, err = st.KeyPut(ctx.Context, &store.Key{
			Name:    name,
			Public:  pub[:],
			Private: privEnc[:],
		})
		if err != nil {
			return err
		}

		return nil
	}

	return cmd
}
