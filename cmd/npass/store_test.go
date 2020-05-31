package main

import (
	"context"
	"testing"
)

func TestNewStore(t *testing.T) {
	st, err := newStore("testing", map[string]string{"mode": "memory", "_secure_delete": "invalid"})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestNewStoreFail(t *testing.T) {
	_, err := newStore("testing", map[string]string{"mode": "memory", "_loc": "invalid"})
	if err == nil {
		t.Fatalf("newStore did not error; want error")
	}
}

func testStore(t *testing.T) store {
	st, err := newStore(t.Name(), map[string]string{"mode": "memory"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	return st
}

func TestStoreInitSchema(t *testing.T) {
	st := testStore(t)

	ok, err := st.checkSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Errorf("checkSchema() = %t; want %t", ok, false)
	}

	err = st.initSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err = st.checkSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("checkSchema() = %t; want %t", ok, true)
	}

	err = st.initSchema(context.Background())
	if err == nil {
		t.Errorf("initSchema() did not error; want error")
	}

	_, err = st.Exec("DELETE FROM meta WHERE key= 'version'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err = st.checkSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Errorf("checkSchema() = %t; want %t", ok, false)
	}
}
