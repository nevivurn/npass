package main

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

type store struct {
	db *sql.DB
}

func newStore(name string) (*store, error) {
	dsn := fmt.Sprintf("file:%s?%s",
		name,
		url.Values(map[string][]string{
			"_foreign_keys":  []string{"1"},
			"_secure_delete": []string{"1"},
		}).Encode(),
	)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return &store{db: db}, nil
}

type storeKey struct {
	name    string
	pub     string
	privEnc string
}

func (st *store) keyList() ([]*storeKey, error) {
	return nil, nil
}
