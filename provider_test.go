package main

import "testing"

func TestClaudeDefaultModel(t *testing.T) {
	c := &claudeProvider{}

	tests := []struct {
		contentType string
		want        string
	}{
		{"code", claudeModelSonnet},
		{"config", claudeModelSonnet},
		{"prompt", claudeModelSonnet},
		{"prose", claudeModelHaiku},
		{"unknown", claudeModelHaiku},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := c.DefaultModel(tt.contentType)
			if got != tt.want {
				t.Errorf("claudeProvider.DefaultModel(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestGeminiDefaultModel(t *testing.T) {
	g := &geminiProvider{}

	tests := []struct {
		contentType string
		want        string
	}{
		{"code", modelPro},
		{"config", modelPro},
		{"prompt", modelPro},
		{"prose", modelFlash},
		{"unknown", modelFlash},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := g.DefaultModel(tt.contentType)
			if got != tt.want {
				t.Errorf("geminiProvider.DefaultModel(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
