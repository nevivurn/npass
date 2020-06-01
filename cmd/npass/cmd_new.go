package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

func (a *app) cmdNew(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errUsage
	}

	key, name, typ, err := parseIdentifier(args[0])
	if err != nil {
		return err
	}

	if key != "" && name != "" && typ != "" {
		return a.cmdNewPass(ctx, key, name, typ)
	}
	if key != "" && name == "" && typ == "" {
		return a.cmdNewKey(ctx, key)
	}

	return errUsage
}

func (a *app) cmdNewKey(ctx context.Context, key string) error {
	// Check if already exists
	var exists bool
	queryExists := `SELECT EXISTS(SELECT 1 FROM keys WHERE name = ?)`
	err := a.st.QueryRowContext(ctx, queryExists, key).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("duplicate key %q", key)
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	pass, err := a.pin.NewPass(ctx, fmt.Sprintf("Enter password for key %q:", key))
	if err != nil {
		return err
	}

	pkey := passKey(pass, salt)
	var keyArr [32]byte
	copy(keyArr[:], pkey)

	privEnc := secretbox.Seal(salt, priv[:], &[24]byte{}, &keyArr)

	queryInsert := `INSERT INTO keys (name, public, private) VALUES(?, ?, ?)`
	_, err = a.st.ExecContext(ctx, queryInsert,
		key,
		base64.RawStdEncoding.EncodeToString(pub[:]),
		base64.RawStdEncoding.EncodeToString(privEnc),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.w, "created new key %q: %s\n", key, base64.RawStdEncoding.EncodeToString(pub[:]))
	return nil
}

func (a *app) cmdNewPass(ctx context.Context, key, name, typ string) error {
	return nil
}
