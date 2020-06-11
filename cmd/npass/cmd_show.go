package main

import (
	"context"
	"database/sql"
	"fmt"
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
	tx, err := a.st.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
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
	panic("not yet implemented")
}

func (a *app) cmdShowName(ctx context.Context, key, name string) error {
	panic("not yet implemented")
}

func (a *app) cmdShowPass(ctx context.Context, key, name, typ string) error {
	panic("not yet implemented")
}
