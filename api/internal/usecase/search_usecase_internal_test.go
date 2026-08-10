//go:build unit

package usecase

import "testing"

func TestBuildPrefixTsQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"single word", "morning", "morning:*"},
		{"multi word", "morning run", "morning:* & run:*"},
		{"tsquery special characters stripped", "a&b|c(d)e:f", "a:* & b:* & c:* & d:* & e:* & f:*"},
		{"leading and trailing whitespace", "  gym  ", "gym:*"},
		{"collapses internal whitespace runs", "a    b", "a:* & b:*"},
		{"punctuation only yields empty string", "!!!", ""},
		{"empty input yields empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPrefixTsQuery(tt.query)
			if got != tt.want {
				t.Errorf("buildPrefixTsQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}
