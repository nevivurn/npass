package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Whether to speed up KDF (for tests).
var fastKDF = false

func passKey(pass string, salt []byte) []byte {
	if fastKDF {
		fmt.Println("WARNING: running with unsafe parameters")
		return argon2.IDKey([]byte(pass), salt, 1, 32, 1, 32)
	}
	return argon2.IDKey([]byte(pass), salt, 1, 6<<20, 8, 32)
}

// Charsets allowed inside identifiers.
const (
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetNumber  = "0123456789"
	charsetAlpha   = charsetLower + charsetUpper
	charsetAlnum   = charsetAlpha + charsetNumber
	charsetSpecial = "!-./?@"

	charsetKey  = charsetLower + charsetNumber + "-"
	charsetName = charsetAlnum + charsetSpecial
	charsetType = charsetLower + "-"
)

var errIdentifier = errors.New("invalid pass identifier")

func parseIdentifier(id string) (key, name, typ string, err error) {
	split := strings.SplitN(id, ":", 3)

	if len(split) >= 1 {
		if len(split[0]) == 0 || !containsOnly(split[0], charsetKey) {
			err = errIdentifier
			return "", "", "", err
		}
		key = split[0]
	}
	if len(split) >= 2 {
		if len(split[1]) == 0 || !containsOnly(split[1], charsetName) {
			err = errIdentifier
			return "", "", "", err
		}
		name = split[1]
	}
	if len(split) >= 3 {
		if len(split[2]) == 0 || !containsOnly(split[2], charsetType) {
			err = errIdentifier
			return "", "", "", err
		}
		typ = split[2]
	}

	return
}

func containsOnly(s string, charset string) bool {
	for _, r := range s {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}
	return true
}
