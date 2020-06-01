package main

import "testing"

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
