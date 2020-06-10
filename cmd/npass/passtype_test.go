package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestPassMap(t *testing.T) {
	_, err := newPass("INVALID")
	if !errors.Is(err, errInvalidPassType) {
		t.Errorf("newPass err = %v; want %v", err, errInvalidPassType)
	}

	p, err := newPass("pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p1, ok := p.(*passPassword); !ok {
		t.Errorf("newPass(%q) = %T; want %T", "pass", p, p1)
	}
}

func TestPassPasswordRead(t *testing.T) {
	pin := &testPinentry{pass: "pass"}
	a, _ := testNewApp(t, pin)
	p := new(passPassword)

	err := p.readPass(context.Background(), a, "testing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(*p) != pin.pass {
		t.Errorf("readPass returned %q; want %q", *p, pin.pass)
	}
}

func TestPassPasswordReadFail(t *testing.T) {
	pin := &testPinentry{err: fmt.Errorf("testing error")}
	a, _ := testNewApp(t, pin)
	p := new(passPassword)

	err := p.readPass(context.Background(), a, "testing")
	if !errors.Is(err, pin.err) {
		t.Errorf("readPass err = %v; want %v", err, errUsage)
	}
}

func TestPassPasswordPrint(t *testing.T) {
	p := new(passPassword)
	*p = passPassword("pass")

	got, err := p.printPass()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != string(*p) {
		t.Errorf("got %q; want %q", got, string(*p))
	}
}

func TestPassPasswordMarshalText(t *testing.T) {
	p := new(passPassword)
	*p = passPassword("pass")

	got, err := p.MarshalText()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, []byte(*p)) {
		t.Errorf("got %v; want %v", got, []byte(*p))
	}
}

func TestPassPasswordUnmarshalText(t *testing.T) {
	p := new(passPassword)
	pass := "pass"

	err := p.UnmarshalText([]byte(pass))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(*p) != pass {
		t.Errorf("got %q; want %q", string(*p), pass)
	}
}
