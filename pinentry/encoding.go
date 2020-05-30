package pinentry

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"strings"
)

func percentEncode(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '%', '\r', '\n':
			fmt.Fprintf(&sb, "%%%02X", r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func percentDecode(s string) (string, error) {
	var (
		sb strings.Builder
		sr = strings.NewReader(s)
	)

	for sr.Len() != 0 {
		r, _, err := sr.ReadRune()
		if err != nil {
			return "", err
		}

		if r != '%' {
			sb.WriteRune(r)
			continue
		}

		pct := ""
		for i := 0; i < 2; i++ {
			r, _, err := sr.ReadRune()
			if err != nil {
				return "", err
			}
			if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') {
				pct += string(r)
			} else {
				return "", fmt.Errorf("invalid percent-encoding")
			}
		}

		dec, err := hex.DecodeString(pct)
		if err != nil {
			return "", err
		}
		sb.Write(dec)
	}

	return sb.String(), nil
}

func send(rw *bufio.ReadWriter, cmd string, args ...string) error {
	eargs := make([]string, len(args)+1)
	eargs[0] = cmd
	for i, arg := range args {
		eargs[i+1] = percentEncode(arg)
	}

	msg := strings.Join(eargs, " ")
	fmt.Fprintln(rw, msg)

	return rw.Flush()
}

func recv(rw *bufio.ReadWriter) (string, error) {
	var resp string
	for {
		rd, err := rw.ReadString('\n')
		if err != nil {
			return "", err
		}
		rd = strings.TrimRight(rd, "\r\n")

		args := strings.SplitN(rd, " ", 2)
		cmd := args[0]

		arg := ""
		if len(args) == 2 {
			arg, err = percentDecode(args[1])
			if err != nil {
				return "", fmt.Errorf("invalid data from pinentry: %w", err)
			}
		}

		switch cmd {
		case "OK":
			return resp, nil
		case "ERR":
			return resp, fmt.Errorf("pinentry error: %s", arg)
		case "D":
			resp = arg
		case "S", "#": // ignored
		default:
			return "", fmt.Errorf("invalid data from pinentry")
		}
	}
}
