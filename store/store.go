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

const (
	sqlVersion = 1
	sqlSchema  = `
CREATE TABLE meta (
	key	TEXT	PRIMARY KEY NOT NULL,
	value	BLOB	NOT NULL
);
INSERT INTO meta (key, value) VALUES('version', 1);

CREATE TABLE key (
	id	INTEGER	PRIMARY KEY NOT NULL,
	name	TEXT	NOT NULL UNIQUE,
	public	TEXT	NOT NULL UNIQUE,
	private	TEXT
);

CREATE TABLE pass (
	id	INTEGER	PRIMARY KEY NOT NULL,
	key_id	INTEGER	NOT NULL REFERENCES key(id),
	name	TEXT	NOT NULL,
	data	TEXT	NOT NULL,
	UNIQUE	(key_id, name)
);
`
)

// Store is the password database.
type Store struct {
	db *sql.DB
}

func openDB(name string, create bool) (*sql.DB, error) {
	q := url.Values{
		"cache":          []string{"shared"},
		"_foreign_keys":  []string{"1"},
		"_secure_delete": []string{"1"},
	}
	if create {
		q.Set("mode", "rwc")
	} else {
		q.Set("mode", "rw")
	}

	dsn := fmt.Sprintf("file:%s?%s", name, q.Encode())

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return db, nil
}

// NewInit creates a new store and initializes the DB schema.
func NewInit(name string) (*Store, error) {
	db, err := openDB(name, true)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(sqlSchema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// New creates a new store backed by the given database.
func New(name string) (*Store, error) {
	db, err := openDB(name, false)
	if err != nil {
		return nil, err
	}

	query := `SELECT value FROM meta WHERE key = 'version'`
	var version int
	if err := db.QueryRow(query).Scan(&version); err != nil {
		return nil, err
	}
	if version != sqlVersion {
		return nil, fmt.Errorf("invalid database version: want %d, got %d", sqlVersion, version)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database connection.
func (st *Store) Close() error {
	return st.db.Close()
}
