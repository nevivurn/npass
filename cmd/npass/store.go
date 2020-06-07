package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

const (
	schema = `
CREATE TABLE keys (
	id	INTEGER	PRIMARY KEY NOT NULL,
	name	TEXT	UNIQUE NOT NULL,
	public	TEXT	NOT NULL,
	private	TEXT
);
CREATE TABLE pass (
	id	INTEGER	PRIMARY KEY NOT NULL,
	key_id	INTEGER	NOT NULL REFERENCES keys(id),
	name	TEXT	NOT NULL,
	type	TEXT	NOT NULL,
	data	TEXT	NOT NULL,
	UNIQUE	(key_id, name, type)
);
CREATE TABLE meta (
	key	TEXT	PRIMARY KEY NOT NULL,
	value	TEXT	NOT NULL
);
INSERT INTO meta (key, value) VALUES('version', ?);
`
	schemaVersion = "1"
)

type store struct{ *sql.DB }

var defaultStoreArgs = map[string]string{
	"_foreign_keys":  "true",
	"_secure_delete": "true",
	"cache":          "shared",
}

func newStore(name string, args map[string]string) (store, error) {
	q := make(url.Values, len(args))
	for k, v := range args {
		q[k] = []string{v}
	}
	for k, v := range defaultStoreArgs {
		q[k] = []string{v}
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?%s", name, q.Encode()))
	if err != nil {
		return store{}, err
	}

	if err := db.Ping(); err != nil {
		return store{}, err
	}

	return store{db}, nil
}

func (st *store) checkSchema(ctx context.Context) (bool, error) {
	queryExists := `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'meta')`

	var exists bool
	err := st.QueryRow(queryExists).Scan(&exists)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	queryVersion := `SELECT value FROM meta WHERE key = 'version'`

	var version string
	err = st.QueryRow(queryVersion).Scan(&version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return version == schemaVersion, nil
}

func (st *store) initSchema(ctx context.Context) error {
	_, err := st.ExecContext(ctx, schema, schemaVersion)
	return err
}
