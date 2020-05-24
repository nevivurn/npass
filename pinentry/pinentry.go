package pinentry

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type pinentry struct {
	prompt  string
	confirm bool
	verify  func(string) bool
}

// Option is used to configure pinentry.
type Option func(*pinentry)

// Prompt sets the password prompt.
func Prompt(prompt string) Option {
	return func(pin *pinentry) {
		pin.prompt = prompt
	}
}

// Confirm makes the user confirm their password.
func Confirm() Option {
	return func(pin *pinentry) {
		pin.confirm = true
	}
}

// Verify specifies the function to check whether the password is correct..
func Verify(f func(string) bool) Option {
	return func(pin *pinentry) {
		pin.verify = f
	}
}

func execPinentry(ctx context.Context) (*exec.Cmd, io.ReadWriteCloser, error) {
	cmd := exec.CommandContext(ctx, "pinentry")

	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	rwc := struct {
		io.Reader
		io.WriteCloser
	}{r, w}

	return cmd, rwc, nil
}

// ReadPassword reads a password from the user.
func ReadPassword(ctx context.Context, opts ...Option) (string, error) {
	pin := pinentry{
		prompt:  "Enter your password",
		confirm: false,
		verify:  func(string) bool { return true },
	}
	for _, opt := range opts {
		opt(&pin)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd, rw, err := execPinentry(ctx)
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	defer cmd.Wait() // nolint:errcheck // error checked on return, this is a fallback in case of errors
	defer cancel()

	brw := bufio.NewReadWriter(bufio.NewReader(rw), bufio.NewWriter(rw))

	// skip initial OK
	if _, err := readResponse(brw); err != nil {
		return "", err
	}

	if _, err := send(brw, "SETDESC", pin.prompt); err != nil {
		return "", err
	}
	if _, err := send(brw, "SETPROMPT", "Password:"); err != nil {
		return "", err
	}
	if pin.confirm {
		if _, err := send(brw, "SETREPEAT", "Confirm:"); err != nil {
			return "", err
		}
		if _, err := send(brw, "SETREPEATERROR", "Passwords do not match"); err != nil {
			return "", err
		}
	}

	// allow running with pinentry-curses
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		tty := os.Stdin.Name()
		// resolve symlink
		for {
			fi, err := os.Lstat(tty)
			if err != nil {
				return "", err
			}

			if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
				break
			}

			tty, err = os.Readlink(tty)
			if err != nil {
				return "", err
			}
		}

		// pinentry does not seem to accept percent-encoding here
		fmt.Fprintf(brw, "OPTION ttyname=%s\n", tty)
		if err := brw.Flush(); err != nil {
			return "", err
		}
		if _, err := readResponse(brw); err != nil {
			return "", err
		}
	}

	var resp string
	for try := 0; try < 3; try++ {
		resps, err := send(brw, "GETPIN")
		if err != nil {
			return "", err
		}

		if len(resps) == 0 {
			if pin.confirm {
				break
			}

			msg := fmt.Sprintf("Password may not be empty (try %d of 3)", try+2)
			if _, err := send(brw, "SETERROR", msg); err != nil {
				return "", err
			}
			continue
		}

		if !pin.verify(resps[0]) {
			if pin.confirm {
				return "", fmt.Errorf("incorrect password")
			}

			msg := fmt.Sprintf("Incorrect password (try %d of 3)", try+2)
			if _, err := send(brw, "SETERROR", msg); err != nil {
				return "", err
			}
			continue
		}

		resp = resps[0]
		break
	}
	if resp == "" {
		return "", fmt.Errorf("could not read password")
	}

	if _, err := send(brw, "BYE"); err != nil {
		return "", err
	}

	return resp, cmd.Wait()
}

func send(rw *bufio.ReadWriter, cmd string, args ...string) ([]string, error) {
	eargs := make([]string, len(args))
	for i, arg := range args {
		eargs[i] = url.PathEscape(arg)
	}

	fmt.Fprintln(rw, cmd, strings.Join(eargs, " "))
	if err := rw.Flush(); err != nil {
		return nil, err
	}

	return readResponse(rw)
}

func readResponse(rw *bufio.ReadWriter) ([]string, error) {
	resp := []string{}
	for {
		rd, err := rw.ReadString('\n')
		if err != nil {
			return resp, err
		}
		rd = strings.TrimSuffix(rd, "\n")

		args := strings.SplitN(rd, " ", 2)
		if len(args) < 1 {
			return resp, fmt.Errorf("invalid data from pinentry")
		}

		for i, arg := range args[1:] {
			args[1+i], err = url.PathUnescape(arg)
			if err != nil {
				return resp, fmt.Errorf("invalid data from pinentry: %w", err)
			}
		}

		switch args[0] {
		case "OK":
			return resp, nil
		case "ERR":
			return resp, fmt.Errorf("pinentry error: %s", strings.Join(args[1:], ""))
		case "D":
			resp = append(resp, args[1:]...)
		case "#":
		}
	}
}
