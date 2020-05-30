// Package pinentry provides an interface to the pinentry utility.
package pinentry

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func execPinentry(ctx context.Context) (*exec.Cmd, *bufio.ReadWriter, error) {
	var cmd *exec.Cmd
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		tty, err := recurReadlink(os.Stdin.Name())
		if err != nil {
			return nil, nil, fmt.Errorf("pinentry: could not find tty")
		}
		cmd = exec.CommandContext(ctx, "pinentry", "--ttyname", tty)
	} else {
		cmd = exec.CommandContext(ctx, "pinentry")
	}

	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	rw := bufio.NewReadWriter(
		bufio.NewReader(r),
		bufio.NewWriter(w),
	)
	return cmd, rw, nil
}

// Confirm displays a confirmation dialog to the user.
func Confirm(ctx context.Context, prompt string) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd, rw, err := execPinentry(ctx)
	if err != nil {
		cancel()
		return false, err
	}
	defer cancel()

	ok, err := confirm(rw, prompt)
	if err != nil {
		return false, err
	}

	if err := cmd.Wait(); err != nil {
		return false, err
	}
	return ok, nil
}

func confirm(rw *bufio.ReadWriter, prompt string) (bool, error) {
	// Skip initial OK
	if _, err := recv(rw); err != nil {
		return false, err
	}

	if err := send(rw, "SETDESC", prompt); err != nil {
		return false, err
	}
	if _, err := recv(rw); err != nil {
		return false, err
	}

	if err := send(rw, "CONFIRM"); err != nil {
		return false, err
	}
	_, err := recv(rw)

	ok := false
	if err != nil && !strings.Contains(err.Error(), "Operation cancelled") {
		return false, err
	}
	if err == nil {
		ok = true
	}

	if err := send(rw, "BYE"); err != nil {
		return false, err
	}
	if _, err := recv(rw); err != nil {
		return false, err
	}

	return ok, nil
}

// NewPass displays a new password creation dialogue.
func NewPass(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd, rw, err := execPinentry(ctx)
	if err != nil {
		cancel()
		return "", err
	}
	defer cancel()

	pass, err := newPass(rw, prompt)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return pass, nil
}

func newPass(rw *bufio.ReadWriter, prompt string) (string, error) {
	// Skip initial OK
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err := send(rw, "SETDESC", prompt); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err := send(rw, "SETPROMPT", "Password:"); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err := send(rw, "SETREPEATERROR", "Passwords do not match"); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	var pass string
	for retry := 3; retry > 0; retry-- {
		if err := send(rw, "SETREPEAT"); err != nil {
			return "", err
		}
		if _, err := recv(rw); err != nil {
			return "", err
		}

		if err := send(rw, "GETPIN"); err != nil {
			return "", err
		}

		var err error
		pass, err = recv(rw)
		if err != nil {
			return "", err
		}
		if pass != "" {
			break
		}

		if err := send(rw, "SETERROR", "The password may not be empty"); err != nil {
			return "", err
		}
		if _, err := recv(rw); err != nil {
			return "", err
		}
	}

	var err error
	if pass == "" {
		err = fmt.Errorf("pinentry: too many retries")
	}

	if err := send(rw, "BYE"); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	return pass, err
}

// AskPass displays a password entry dialogue.
func AskPass(ctx context.Context, prompt string, verify func(string) bool) (string, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd, rw, err := execPinentry(ctx)
	if err != nil {
		cancel()
		return "", err
	}
	defer cancel()

	pass, err := askPass(rw, prompt, verify)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return pass, nil
}

func askPass(rw *bufio.ReadWriter, prompt string, verify func(string) bool) (string, error) {
	// Skip initial OK
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err := send(rw, "SETDESC", prompt); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err := send(rw, "SETPROMPT", "Password:"); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	var (
		pass string
		ok   bool
	)
	for retry := 3; retry > 0; retry-- {
		if err := send(rw, "GETPIN"); err != nil {
			return "", err
		}

		var err error
		pass, err = recv(rw)
		if err != nil {
			return "", err
		}
		if ok = verify(pass); ok {
			break
		}

		if err := send(rw, "SETERROR", "Incorrect password"); err != nil {
			return "", err
		}
		if _, err := recv(rw); err != nil {
			return "", err
		}
	}

	var err error
	if !ok {
		err = fmt.Errorf("pinentry: too many retries")
	}

	if err := send(rw, "BYE"); err != nil {
		return "", err
	}
	if _, err := recv(rw); err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}
	return pass, nil
}

func recurReadlink(name string) (string, error) {
	for {
		fi, err := os.Lstat(name)
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
			break
		}

		name, err = os.Readlink(name)
		if err != nil {
			return "", err
		}
	}

	return name, nil
}
