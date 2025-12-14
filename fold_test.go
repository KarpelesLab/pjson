// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pjson

import (
	"bytes"
	"testing"
)

func TestFoldName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a", "A"},
		{"A", "A"},
		{"abc", "ABC"},
		{"ABC", "ABC"},
		{"AbC", "ABC"},
		{"hello", "HELLO"},
		{"HELLO", "HELLO"},
		{"Hello", "HELLO"},
	}

	for _, tt := range tests {
		got := string(foldName([]byte(tt.input)))
		if got != tt.want {
			t.Errorf("foldName(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestFoldNameEquality(t *testing.T) {
	// Test that foldName produces the same result for case-insensitive matches
	pairs := []struct {
		s1, s2 string
		equal  bool
	}{
		{"a", "A", true},
		{"abc", "ABC", true},
		{"AbC", "aBC", true},
		{"a", "b", false},
		{"abc", "abd", false},
	}

	for _, tt := range pairs {
		f1 := foldName([]byte(tt.s1))
		f2 := foldName([]byte(tt.s2))
		got := bytes.Equal(f1, f2)
		if got != tt.equal {
			t.Errorf("foldName(%q) == foldName(%q) = %v; want %v", tt.s1, tt.s2, got, tt.equal)
		}
	}
}
