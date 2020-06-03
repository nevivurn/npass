package main

import (
	"context"
	"encoding"
	"errors"
	"fmt"
)

var errInvalidPassType = errors.New("invalid pass type")

type passType interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler

	readPass(context.Context, *app, string) error
	printPass() (string, error)
	printRawPass() (string, error)
}

var passTypeMap = map[string]func() passType{
	"pass": func() passType { return new(passPassword) },
}

func newPass(typ string) (passType, error) {
	f, ok := passTypeMap[typ]
	if !ok {
		return nil, errInvalidPassType
	}
	return f(), nil
}

type passPassword string

func (p *passPassword) readPass(ctx context.Context, a *app, name string) error {
	pass, err := a.pin.NewPass(ctx, fmt.Sprintf("Enter password for %q:", name))
	if err != nil {
		return err
	}

	*p = passPassword(pass)
	return nil
}
func (p *passPassword) printPass() (string, error) {
	return string(*p), nil
}

func (p *passPassword) printRawPass() (string, error) {
	return p.printPass()
}

func (p *passPassword) MarshalText() ([]byte, error) {
	return []byte(*p), nil
}

func (p *passPassword) UnmarshalText(b []byte) error {
	*p = passPassword(b)
	return nil
}
