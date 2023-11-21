package labeler

import (
	"testing"
	"time"
)

func TestParseExtendedDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", 1 * time.Second},
		{"2m", 2 * time.Minute},
		{"3h", 3 * time.Hour},
		{"4d", 4 * 24 * time.Hour},
		{"5w", 5 * 7 * 24 * time.Hour},
		{"6y", 6 * 365 * 24 * time.Hour},
	}

	for _, test := range tests {
		result, err := parseExtendedDuration(test.input)
		if err != nil {
			t.Errorf("failed to parse duration from %s: %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}
