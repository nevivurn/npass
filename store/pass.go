package store

import (
	"context"
	"database/sql"
	"encoding/base64"
)

// Pass is a password stored in the database.
type Pass struct {
	ID    int64
	KeyID int64
	Name  string
	Data  []byte
}

func (p *Pass) scan(row scanner) error {
	var data64 string
	err := row.Scan(&p.ID, &p.KeyID, &p.Name, &data64)
	if err != nil {
		return err
	}

	p.Data, err = base64.RawStdEncoding.DecodeString(data64)
	if err != nil {
		return err
	}

	return nil
}

// PassList returns the list of passwords stored in the database.
func (st *Store) PassList(ctx context.Context) ([]*Pass, error) {
	query := `SELECT id, key_id, name, data FROM pass`
	rows, err := st.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pass []*Pass
	for rows.Next() {
		p := &Pass{}
		if err := p.scan(rows); err != nil {
			return nil, err
		}
		pass = append(pass, p)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return pass, nil
}

// PassListKey returns the list of passwords encrypted with the given key
// stored in the database.
func (st *Store) PassListKey(ctx context.Context, keyID int64) ([]*Pass, error) {
	query := `SELECT id, key_id, name, data FROM pass WHERE key_id = ?`
	rows, err := st.db.QueryContext(ctx, query, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pass []*Pass
	for rows.Next() {
		p := &Pass{}
		if err := p.scan(rows); err != nil {
			return nil, err
		}
		pass = append(pass, p)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return pass, nil
}

// PassListName returns the list of passwords with the given name stored in
// the database.
func (st *Store) PassListName(ctx context.Context, name string) ([]*Pass, error) {
	query := `SELECT id, key_id, name, data FROM pass WHERE name = ?`
	rows, err := st.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pass []*Pass
	for rows.Next() {
		p := &Pass{}
		if err := p.scan(rows); err != nil {
			return nil, err
		}
		pass = append(pass, p)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return pass, nil
}

// PassFindID finds the password identified by the given id.
func (st *Store) PassFindID(ctx context.Context, id int64) (*Pass, error) {
	query := `SELECT id, key_id, name, data FROM key WHERE id = ? LIMIT 1`
	row := st.db.QueryRowContext(ctx, query, id)

	p := &Pass{}
	if err := p.scan(row); err != nil {
		return nil, err
	}

	return p, nil
}

// PassPut inserts a new password into the database.
func (st *Store) PassPut(ctx context.Context, p *Pass) (int64, error) {
	query := `INSERT INTO pass (id, key_id, name, data) VALUES (?, ?, ?, ?)`

	data64 := base64.RawStdEncoding.EncodeToString(p.Data)

	res, err := st.db.ExecContext(ctx, query, p.ID, p.KeyID, p.Name, data64)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// PassDel deletes a key from the database.
func (st *Store) PassDel(ctx context.Context, id int64) error {
	query := `DELETE FROM pass WHERE id = ?`

	res, err := st.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return sql.ErrNoRows
	}

	return nil
}
