package main

import (
	"testing"
)

func TestParseClassification(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		// Uppercase inputs — exact match
		{name: "uppercase CODE", raw: "CODE", want: "code"},
		{name: "uppercase CONFIG", raw: "CONFIG", want: "config"},
		{name: "uppercase PROSE", raw: "PROSE", want: "prose"},

		// Whitespace trimming
		{name: "leading/trailing spaces", raw: "  CODE  ", want: "code"},
		{name: "newline wrapped", raw: "\nPROSE\n", want: "prose"},

		// Mixed case — uppercased before matching
		{name: "title case Code", raw: "Code", want: "code"},
		{name: "mixed case cOnFiG", raw: "cOnFiG", want: "config"},
		{name: "lowercase code", raw: "code", want: "code"},

		// Invalid / unknown → default to prose
		{name: "invalid INVALID", raw: "INVALID", want: "prose"},
		{name: "unknown json", raw: "json", want: "prose"},
		{name: "empty string", raw: "", want: "prose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClassification(tt.raw)
			if got != tt.want {
				t.Errorf("parseClassification(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestResolveKeyFromEnv(t *testing.T) {
	const envVar = "TEST_WRITEBETTERER_KEY"
	const opRef = "op://vault/item/field"

	t.Run("plain key from env", func(t *testing.T) {
		t.Setenv(envVar, "sk-test-key-123")

		got, err := resolveKey(envVar, opRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "sk-test-key-123" {
			t.Errorf("resolveKey() = %q, want %q", got, "sk-test-key-123")
		}
	})

	t.Run("empty env falls through to opRead", func(t *testing.T) {
		t.Setenv(envVar, "")

		_, err := resolveKey(envVar, opRef)
		if err == nil {
			t.Fatal("expected error when op CLI is unavailable, got nil")
		}
	})
}
