package store

import (
	"context"
	"database/sql"
	"encoding/base64"
)

// Key is an encryption keypair stored in the database.
type Key struct {
	ID      int64
	Name    string
	Public  []byte
	Private []byte
}

func (k *Key) scan(row scanner) error {
	var pub64, priv64 string
	err := row.Scan(&k.ID, &k.Name, &pub64, &priv64)
	if err != nil {
		return err
	}

	k.Public, err = base64.RawStdEncoding.DecodeString(pub64)
	if err != nil {
		return err
	}
	k.Private, err = base64.RawStdEncoding.DecodeString(priv64)
	if err != nil {
		return err
	}

	return nil
}

// KeyList returns the list of keys stored in the database.
func (st *Store) KeyList(ctx context.Context) ([]*Key, error) {
	query := `SELECT id, name, public, private FROM key ORDER BY id DESC`
	rows, err := st.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*Key
	for rows.Next() {
		k := &Key{}
		if err := k.scan(rows); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return keys, nil
}

// KeyFindID finds the key identified by the given id.
func (st *Store) KeyFindID(ctx context.Context, id int64) (*Key, error) {
	query := `SELECT id, name, public, private FROM key WHERE id = ? LIMIT 1`
	row := st.db.QueryRowContext(ctx, query, id)

	k := &Key{}
	if err := k.scan(row); err != nil {
		return nil, err
	}

	return k, nil
}

// KeyFindName finds the key identified by the given name.
func (st *Store) KeyFindName(ctx context.Context, name string) (*Key, error) {
	query := `SELECT id, name, public, private FROM key WHERE name = ? LIMIT 1`
	row := st.db.QueryRowContext(ctx, query, name)

	k := &Key{}
	if err := k.scan(row); err != nil {
		return nil, err
	}

	return k, nil
}

// KeyFindPublic finds the key identified by the given public key.
func (st *Store) KeyFindPublic(ctx context.Context, public []byte) (*Key, error) {
	query := `SELECT id, name, public, private FROM key WHERE public = ? LIMIT 1`

	qpub64 := base64.RawStdEncoding.EncodeToString(public)
	row := st.db.QueryRowContext(ctx, query, qpub64)

	k := &Key{}
	if err := k.scan(row); err != nil {
		return nil, err
	}

	return k, nil
}

// KeyPut inserts a new key into the database.
func (st *Store) KeyPut(ctx context.Context, k *Key) (int64, error) {
	query := `INSERT INTO key (name, public, private) VALUES (?, ?, ?)`

	pub64 := base64.RawStdEncoding.EncodeToString(k.Public)
	priv64 := base64.RawStdEncoding.EncodeToString(k.Private)

	res, err := st.db.ExecContext(ctx, query, k.Name, pub64, priv64)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// KeyDel deletes a key from the database.
func (st *Store) KeyDel(ctx context.Context, id int64) error {
	query := `DELETE FROM key WHERE id = ?`

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
