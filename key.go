package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	// ref: https://www.ietf.org/id/draft-irtf-cfrg-argon2-10.txt section 4
	// tuned for the author's machine
	return argon2.IDKey(pass, salt, 1, 8<<20, 12, 32)
}

func cmdKey() *cli.Command {
	cmd := &cli.Command{
		Name: "key",

		Subcommands: []*cli.Command{
			cmdKeyGet(),
			cmdKeyPut(),
			cmdKeyDel(),
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
		if err == nil {
			return fmt.Errorf("key %q already exists", name)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		pub, priv, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		pass, err := pinentry.ReadPassword(
			ctx.Context,
			pinentry.Prompt(fmt.Sprintf("Enter password for key %q", name)),
			pinentry.Confirm(),
		)
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

func cmdKeyDel() *cli.Command {
	cmd := &cli.Command{
		Name: "del",

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

		key, err := st.KeyFindName(ctx.Context, name)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("key %q does not exist", name)
		}
		if err != nil {
			return err
		}

		err = st.KeyDel(ctx.Context, key.ID)
		if err != nil {
			return err
		}

		return nil
	}

	return cmd
}

func cmdKeyGet() *cli.Command {
	cmd := &cli.Command{
		Name: "get",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: flagKeyName,
			},
		},
	}

	cmd.Action = func(ctx *cli.Context) error {
		st, err := store.New(ctx.Path(flagDB))
		if err != nil {
			return err
		}
		defer st.Close()

		var keys []*store.Key
		if ctx.IsSet(flagKeyName) {
			name := ctx.String(flagKeyName)
			key, err := st.KeyFindName(ctx.Context, name)
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("key %q does not exist", name)
			}
			if err != nil {
				return err
			}
			keys = append(keys, key)
		} else {
			keys, err = st.KeyList(ctx.Context)
			if err != nil {
				return err
			}
		}

		for _, k := range keys {
			fmt.Println(k.Name, base64.RawStdEncoding.EncodeToString(k.Public))
		}

		return nil
	}

	return cmd
}
