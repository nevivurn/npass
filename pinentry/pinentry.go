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
func ReadPassword(ctx context.Context, prompt string) (string, error) {
	return ReadPasswordVerify(ctx, prompt, func(string) bool { return true })
}

// ReadPasswordVerify reads a password from the user, using the given verify function
// to retry.
func ReadPasswordVerify(ctx context.Context, prompt string, verify func(string) bool) (string, error) {
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

	if prompt == "" {
		prompt = "Enter password"
	}
	if _, err := send(brw, "SETDESC", prompt); err != nil {
		return "", err
	}
	if _, err := send(brw, "SETPROMPT", "Password:"); err != nil {
		return "", err
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
			msg := fmt.Sprintf("Password may not be empty (try %d of 3)", try+2)
			if _, err := send(brw, "SETERROR", msg); err != nil {
				return "", err
			}
			continue
		}

		if !verify(resps[0]) {
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
		return "", fmt.Errorf("too many retries")
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
