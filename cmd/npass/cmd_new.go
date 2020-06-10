package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

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
	fullName := strings.Join([]string{key, name, typ}, ":")

	var (
		keyID  int64
		keyPub string
	)
	queryKey := `SELECT id, public FROM keys WHERE name = ? LIMIT 1`
	err := a.st.QueryRowContext(ctx, queryKey, key).Scan(&keyID, &keyPub)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("non-existent key %q", key)
	}
	if err != nil {
		return err
	}

	ptype, err := newPass(typ)
	if err != nil {
		return err
	}

	var exists bool
	queryExists := `SELECT EXISTS(SELECT 1 FROM pass WHERE key_id = ? AND name = ? AND type = ?)`
	err = a.st.QueryRowContext(ctx, queryExists, keyID, name, typ).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("duplicate pass %q", fullName)
	}

	err = ptype.readPass(ctx, a, name)
	if err != nil {
		return err
	}

	passData, err := ptype.MarshalText()
	if err != nil {
		return err
	}
	passData = bytes.Join([][]byte{[]byte(fullName), passData}, []byte(":"))

	keyPubRaw, err := base64.RawStdEncoding.DecodeString(keyPub)
	if err != nil {
		return err
	}

	var keyPubArr [32]byte
	copy(keyPubArr[:], keyPubRaw)

	passEnc, err := box.SealAnonymous(nil, passData, &keyPubArr, rand.Reader)
	if err != nil {
		return err
	}

	queryInsert := `INSERT INTO pass (key_id, name, type, data) VALUES(?, ?, ?, ?)`
	_, err = a.st.ExecContext(ctx, queryInsert,
		keyID, name, typ,
		base64.RawStdEncoding.EncodeToString(passEnc),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.w, "created new pass %q\n", fullName)
	return nil
}
