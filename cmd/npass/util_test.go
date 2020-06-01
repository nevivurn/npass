package main

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestPassKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	got := passKey(
		"0123456789abcdef",
		[]byte("0123456789abcdef"),
	)
	// Obtained with argon2-cffi python library
	want, _ := hex.DecodeString("324fc34ab73bd55a748fbe25dc4c122080fa968c82ac0b19ee67285993fa64eb")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("passKey() = %v; want %v", got, want)
	}
}

func TestPassKeyFast(t *testing.T) {
	fastKDF = true
	defer func() { fastKDF = false }()

	got := passKey(
		"0123456789abcdef",
		[]byte("0123456789abcdef"),
	)
	// Obtained with argon2-cffi python library
	want, _ := hex.DecodeString("92c2708b3e6e914ce3a440f0e2851318c5c400edb0b2c3689d42a1b60f9bdf51")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("passKey() = %v; want %v", got, want)
	}
}

func BenchmarkPassKey(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}
	b.ReportAllocs()

	pass := "0123456789abcdef"
	salt := []byte("0123456789abcdef")
	for i := 0; i < b.N; i++ {
		passKey(pass, salt)
	}
}

func TestParseIdentifier(t *testing.T) {
	type testCase struct {
		key, name, typ string
		err            error
	}
	tests := map[string]testCase{
		"key":          {"key", "", "", nil},
		"key:name":     {"key", "name", "", nil},
		"key:name:typ": {"key", "name", "typ", nil},

		"":          {"", "", "", errIdentifier},
		"KEY":       {"", "", "", errIdentifier},
		"::::":      {"", "", "", errIdentifier},
		"key:":      {"", "", "", errIdentifier},
		"key:::":    {"", "", "", errIdentifier},
		"key:name:": {"", "", "", errIdentifier},
	}

	for tc, want := range tests {
		key, name, typ, err := parseIdentifier(tc)
		if got := (testCase{key, name, typ, err}); got != want {
			t.Errorf("parseIdentifier(%q) = %#v; want %#v", tc, got, want)
		}
	}
}

func TestContainsOnly(t *testing.T) {
	type testCase struct {
		s, cs string
	}
	tests := map[testCase]bool{
		{}:         true,
		{"a", "a"}: true,
		{"a", "b"}: false,
		{"a", ""}:  false,
		{"", "a"}:  true,
		{"HelloWorld!", charsetAlpha + charsetSpecial}: true,
		{"helloWorld!", charsetAlpha}:                  false,
	}

	for tc, want := range tests {
		got := containsOnly(tc.s, tc.cs)
		if got != want {
			t.Errorf("containsOnly(%#v) = %t; want %t", tc, got, want)
		}
	}
}
