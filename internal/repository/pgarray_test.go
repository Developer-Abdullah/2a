package repository

import "testing"

func TestEncodePGTextArray(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", nil, "{}"},
		{"single", []string{"alpha"}, `{"alpha"}`},
		{"multiple", []string{"a", "b", "c"}, `{"a","b","c"}`},
		{"with space", []string{"Fast sync"}, `{"Fast sync"}`},
		{"with quote", []string{`he"llo`}, `{"he\"llo"}`},
		{"with backslash", []string{`a\b`}, `{"a\\b"}`},
	}
	for _, c := range cases {
		if got := encodePGTextArray(c.in); got != c.want {
			t.Errorf("%s: encodePGTextArray(%v) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}
