package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRunFunc(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, 0)
	args := []string{"1", "2", "3"}
	err := errors.New("test error")

	fn := func(ctx1 context.Context, args1 []string) error {
		if ctx1 != ctx {
			t.Errorf("ctx mismatch")
		}
		if !reflect.DeepEqual(args1, args) {
			t.Errorf("args mismatch: got %#v; want %#v", args1, args)
		}
		return err
	}

	err1 := runFunc(fn).run(ctx, args)
	if err1 != err {
		t.Errorf("error mismatch: got %#v; want %#v", err1, err)
	}
}

func TestRunMap(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, 0)
	args := []string{"arg", "1", "2", "3"}
	err := errors.New("test error")

	fn := func(ctx1 context.Context, args1 []string) error {
		if ctx1 != ctx {
			t.Errorf("ctx mismatch")
		}
		if !reflect.DeepEqual(args1, args[1:]) {
			t.Errorf("args mismatch: got %#v; want %#v", args1, args[1:])
		}
		return err
	}

	rm := runMap{args[0]: runFunc(fn)}

	if err := rm.run(ctx, nil); err != errUsage {
		t.Errorf("error mismatch: got %#v; want %#v", err, errUsage)
	}

	if err := rm.run(ctx, []string{"invalid"}); err != errUsage {
		t.Errorf("error mismatch: got %#v; want %#v", err, errUsage)
	}

	err1 := rm.run(ctx, args)
	if err1 != err {
		t.Errorf("error mismatch: got %#v; want %#v", err1, err)
	}
}
