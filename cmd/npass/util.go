package main

import (
	"errors"
	"strings"
)

// Charsets allowed inside identifiers.
const (
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetNumber  = "0123456789"
	charsetAlpha   = charsetLower + charsetUpper
	charsetAlnum   = charsetAlpha + charsetNumber
	charsetSpecial = "!-./?@"
)

var errIdentifier = errors.New("invalid pass identifier")

func parseIdentifier(id string) (key, name, typ string, err error) {
	split := strings.SplitN(id, ":", 3)

	if len(split) >= 1 {
		if len(split[0]) == 0 || !containsOnly(split[0], charsetLower) {
			err = errIdentifier
			return "", "", "", err
		}
		key = split[0]
	}
	if len(split) >= 2 {
		if len(split[1]) == 0 || !containsOnly(split[1], charsetAlnum+charsetSpecial) {
			err = errIdentifier
			return "", "", "", err
		}
		name = split[1]
	}
	if len(split) >= 3 {
		if len(split[2]) == 0 || !containsOnly(split[2], charsetLower) {
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
