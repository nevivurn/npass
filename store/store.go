package store

import (
	"database/sql"
	"fmt"
	"net/url"

	// register sqlite3 sql driver
	_ "github.com/mattn/go-sqlite3"
)

type scanner interface {
	Scan(...interface{}) error
}

const sqlSchema = `
CREATE TABLE IF NOT EXISTS key (
	id	INTEGER	PRIMARY KEY NOT NULL,
	name	TEXT	NOT NULL UNIQUE,
	public	TEXT	NOT NULL UNIQUE,
	private	TEXT);

CREATE TABLE IF NOT EXISTS pass (
	id	INTEGER	PRIMARY KEY NOT NULL,
	key_id	INTEGER	NOT NULL REFERENCES key(id),
	name	TEXT	NOT NULL,
	data	TEXT	NOT NULL,
	UNIQUE	(key_id, name)
);
`

// Store is the password database.
type Store struct {
	db *sql.DB
}

func openDB(name string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?%s", name, url.Values{
		"cache":          []string{"shared"},
		"_foreign_keys":  []string{"1"},
		"_secure_delete": []string{"1"},
	})

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return db, nil
}

// New creates a new store backed by the given database.
func New(name string) (*Store, error) {
	db, err := openDB(name)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(sqlSchema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database connection.
func (st *Store) Close() error {
	return st.db.Close()
}
