package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

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
	tx, err := a.st.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	queryKeys := `SELECT id, name, public FROM keys ORDER BY name`
	rows, err := tx.Query(queryKeys)
	if err != nil {
		return err
	}
	defer rows.Close()

	type dbKey struct {
		id        int64
		name, pub string
	}

	var keys []dbKey
	for rows.Next() {
		var k dbKey
		err := rows.Scan(&k.id, &k.name, &k.pub)
		if err != nil {
			return err
		}
		keys = append(keys, k)
	}

	queryPass := `SELECT key_id, name, type FROM pass ORDER BY key_id, name, type`
	rows, err = tx.Query(queryPass)
	if err != nil {
		return err
	}
	defer rows.Close()

	type dbPass struct {
		name, typ string
	}

	pass := make(map[int64][]dbPass)
	for rows.Next() {
		var (
			p   dbPass
			kid int64
		)
		err := rows.Scan(&kid, &p.name, &p.typ)
		if err != nil {
			return err
		}
		pass[kid] = append(pass[kid], p)
	}

	for _, k := range keys {
		fmt.Fprintf(a.w, "%s %s:\n", k.name, k.pub)
		for _, p := range pass[k.id] {
			fmt.Fprintf(a.w, "  %s: [%s]\n", p.name, p.typ)
		}
	}

	return tx.Commit()
}

func (a *app) cmdShowKey(ctx context.Context, key string) error {
	tx, err := a.st.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		kid  int64
		kpub string
	)
	queryKey := `SELECT id, public FROM keys WHERE name = ? LIMIT 1`
	err = tx.QueryRow(queryKey, key).Scan(&kid, &kpub)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("non-existent key %q", key)
	}
	if err != nil {
		return err
	}

	queryPass := `SELECT name, type FROM pass WHERE key_id = ? ORDER BY name, type`
	rows, err := tx.Query(queryPass, kid)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Fprintf(a.w, "%s %s:\n", key, kpub)

	for rows.Next() {
		var name, typ string
		err := rows.Scan(&name, &typ)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.w, "  %s: [%s]\n", name, typ)
	}

	return tx.Commit()
}

func (a *app) cmdShowName(ctx context.Context, key, name string) error {
	tx, err := a.st.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var kid int64
	queryKey := `SELECT id FROM keys WHERE name = ? LIMIT 1`
	err = tx.QueryRow(queryKey, key).Scan(&kid)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("non-existent key %q", key)
	}
	if err != nil {
		return err
	}

	queryPass := `SELECT type FROM pass WHERE key_id = ? AND name = ? ORDER BY type`
	rows, err := tx.Query(queryPass, kid, name)
	if err != nil {
		return err
	}
	defer rows.Close()

	exists := false
	for rows.Next() {
		exists = true

		var typ string
		err := rows.Scan(&typ)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.w, "%s:%s:%s\n", key, name, typ)
	}

	if !exists {
		return fmt.Errorf("non-existent pass %q", fmt.Sprintf("%s:%s", key, name))
	}

	return tx.Commit()
}

func (a *app) cmdShowPass(ctx context.Context, key, name, typ string) error {
	tx, err := a.st.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		kid             int64
		keyPub, keyPriv string
	)
	queryKey := `SELECT id, public, private FROM keys WHERE name = ? LIMIT 1`
	err = tx.QueryRow(queryKey, key).Scan(&kid, &keyPub, &keyPriv)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("non-existent key %q", key)
	}
	if err != nil {
		return err
	}

	keyPubRaw, err := base64.RawStdEncoding.DecodeString(keyPub)
	if err != nil {
		return err
	}
	keyPrivRaw, err := base64.RawStdEncoding.DecodeString(keyPriv)
	if err != nil {
		return err
	}

	var exists bool
	queryNameExists := `SELECT EXISTS(SELECT 1 FROM pass WHERE key_id = ? AND name = ?)`
	err = tx.QueryRow(queryNameExists, kid, name).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("non-existent pass %q", fmt.Sprintf("%s:%s", key, name))
	}

	pass, err := newPass(typ)
	if err != nil {
		return err
	}

	var passData []byte
	queryPass := `SELECT data FROM pass WHERE key_id = ? AND name = ? AND type = ? LIMIT 1`
	err = tx.QueryRow(queryPass, kid, name, typ).Scan(&passData)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("non-existent pass %s",
			fmt.Sprintf("%s:%s:%s", key, name, typ))
	}
	if err != nil {
		return err
	}

	passDataDec, err := base64.RawStdEncoding.DecodeString(string(passData))
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	keyPrivRaw = keyPrivRaw[copy(salt, keyPrivRaw):]

	var keyPrivDec []byte
	_, err = a.pin.AskPass(ctx, fmt.Sprintf("Enter password for key %q:", key),
		func(pass string) bool {
			pkey := passKey(pass, salt)
			var keyArr [32]byte
			copy(keyArr[:], pkey)

			dec, ok := secretbox.Open(nil, keyPrivRaw, &[24]byte{}, &keyArr)
			if !ok {
				return false
			}

			keyPrivDec = dec
			return true
		},
	)
	if err != nil {
		return err
	}

	var keyPubArr, keyPrivArr [32]byte
	copy(keyPubArr[:], keyPubRaw)
	copy(keyPrivArr[:], keyPrivDec)

	passDec, ok := box.OpenAnonymous(nil, passDataDec, &keyPubArr, &keyPrivArr)
	if !ok {
		return fmt.Errorf("decryption error")
	}

	if err := pass.UnmarshalText(passDec); err != nil {
		return err
	}

	out, err := pass.printPass()
	if err != nil {
		return err
	}

	split := strings.SplitN(out, ":", 4)
	if len(split) != 4 || split[0] != key || split[1] != name || split[2] != typ {
		return fmt.Errorf("decryption error")
	}
	fmt.Fprintln(a.w, split[3])

	return tx.Commit()
}
