package pinentry

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testPipe(t *testing.T) (*bufio.ReadWriter, *bufio.ReadWriter) {
	aR, bW := io.Pipe()
	bR, aW := io.Pipe()

	t.Cleanup(func() {
		aR.Close()
		aW.Close()
		bR.Close()
		bW.Close()
	})

	a := bufio.NewReadWriter(
		bufio.NewReader(aR),
		bufio.NewWriter(aW),
	)
	b := bufio.NewReadWriter(
		bufio.NewReader(bR),
		bufio.NewWriter(bW),
	)
	return a, b
}

func TestConfirm(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		ok, err := confirm(aio, "prompt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !ok {
			t.Errorf("confirm() = %t; want %t", ok, true)
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "CONFIRM\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestConfirmCancel(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		ok, err := confirm(aio, "prompt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if ok {
			t.Errorf("confirm() = %t; want %t", ok, false)
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "CONFIRM\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "ERR", "83886179 Operation cancelled <Pinentry>")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestNewPass(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		pass, err := newPass(aio, "prompt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if pass != "pass" {
			t.Errorf("newPass() = %q; want %q", pass, "pass")
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETPROMPT Password:\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEATERROR Passwords do not match\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEAT\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR The password may not be empty\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEAT\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "D pass")
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestNewPassRetry(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		_, err := newPass(aio, "prompt")
		if want := fmt.Errorf("pinentry: too many retries"); !reflect.DeepEqual(err, want) {
			t.Errorf("newPass() = %v; want %v", err, want)
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETPROMPT Password:\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEATERROR Passwords do not match\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEAT\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR The password may not be empty\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEAT\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR The password may not be empty\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETREPEAT\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR The password may not be empty\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestAskPass(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		verify := func(s string) bool {
			return s == "pass"
		}
		pass, err := askPass(aio, "prompt", verify)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if pass != "pass" {
			t.Errorf("askPass() = %q; want %q", pass, "pass")
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETPROMPT Password:\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR Incorrect password\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "D pass")
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestAskPassRetry(t *testing.T) {
	aio, bio := testPipe(t)

	done := make(chan struct{})
	go func() {
		defer close(done)

		verify := func(s string) bool {
			return s == "pass"
		}
		_, err := askPass(aio, "prompt", verify)
		if want := fmt.Errorf("pinentry: too many retries"); !reflect.DeepEqual(err, want) {
			t.Errorf("askPass() = %v; want %v", err, want)
		}
	}()

	_ = send(bio, "OK", "Pleased to meet you")

	resp, _ := bio.ReadString('\n')
	if want := "SETDESC prompt\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETPROMPT Password:\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "D incorrect")
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR Incorrect password\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "D incorrect")
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR Incorrect password\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "GETPIN\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "D incorrect")
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "SETERROR Incorrect password\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	resp, _ = bio.ReadString('\n')
	if want := "BYE\n"; resp != want {
		t.Fatalf("got %q; want %q", resp, want)
	}
	_ = send(bio, "OK")

	<-done
}

func TestRecurReadlink(t *testing.T) {
	tmp, err := ioutil.TempDir("", "pinentry-test-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmp)

	_, err = recurReadlink(filepath.Join(tmp, "not-found"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("recurReadlink() = %v; want %v", err, os.ErrNotExist)
	}

	f, err := os.Create(filepath.Join(tmp, "0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Close()

	for i := 1; i < 5; i++ {
		err := os.Symlink(filepath.Join(tmp, fmt.Sprint(i-1)), filepath.Join(tmp, fmt.Sprint(i)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	name, err := recurReadlink(filepath.Join(tmp, "4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := filepath.Join(tmp, "0"); name != want {
		t.Errorf("recurReadlink() = %q; want %q", name, want)
	}
}
