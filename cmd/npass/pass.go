package main

import (
	"fmt"
	"strings"
)

type passName struct {
	key  string
	name string
	typ  string
}

const (
	// Various acceptable character sets
	charsetLower  = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetNumber = "0123456789"
	charsetAlpha  = charsetLower + charsetUpper
	charsetAlnum  = charsetAlpha + charsetNumber

	charsetKey  = charsetLower + charsetNumber + "-_"
	charsetName = charsetAlnum + "!-.@_"
	charsetTyp  = charsetLower
)

func parsePass(s string) (p passName, err error) {
	split := strings.SplitN(s, ":", 3)
	if len(s) == 0 || len(split) == 0 {
		err = fmt.Errorf("invalid pass identifier %q", s)
		return
	}

	if len(split) == 3 {
		if split[2] == "" || !containsOnly(split[2], charsetTyp) {
			err = fmt.Errorf("invalid pass identifier %q", s)
			return
		}
		p.typ = split[2]
	}
	if len(split) >= 2 {
		if split[1] == "" || !containsOnly(split[1], charsetName) {
			err = fmt.Errorf("invalid pass identifier %q", s)
			return
		}
		p.name = split[1]
	}
	if len(split) >= 1 {
		if split[0] == "" || !containsOnly(split[0], charsetKey) {
			err = fmt.Errorf("invalid pass identifier %q", s)
			return
		}
		p.key = split[0]
	}
	return
}

func containsOnly(s, charset string) bool {
	for _, r := range s {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}
	return true
}
