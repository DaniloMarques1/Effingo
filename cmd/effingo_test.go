package cmd

import (
	"testing"
	"time"
)

func TestHashCachExpired(t *testing.T) {
	cases := []struct {
		label    string
		input    int64
		expected bool
	}{
		{"Should return true", time.Now().Unix() - 500, true},
		{"Should return false", time.Now().Unix(), false},
		{"Should return false", time.Now().Unix() - 119, false},
		{"Should return true", time.Now().Unix() - 780, true},
		{"Should return true", time.Now().Unix() - 121, true},
		{"Should return false", time.Now().Unix() - 120, false},
	}

	traverser := &DirTraverser{}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			output := traverser.hasCacheExpired(tc.input)
			if output != tc.expected {
				t.Fatalf("Wrong output. expected %v got %v", tc.expected, output)
			}
		})

	}
}
