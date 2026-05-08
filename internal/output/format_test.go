package output

import (
	"strings"
	"testing"
)

func TestFormatEpochSeconds(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string // partial match ok
	}{
		{"int64", int64(1700000000), "2023"},   // Tue Nov 14 2023 22:13:20 UTC
		{"int", 1700000000, "2023"},
		{"float64", float64(1700000000), "2023"},
		{"string-numeric", "1700000000", "2023"},
		{"zero", int64(0), ""},
		{"unparseable string", "not-a-number", ""},
		{"nil-like", any(nil), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatEpochSeconds(tc.in)
			if tc.want == "" && got != "" {
				t.Errorf("expected empty for %v; got %q", tc.in, got)
			}
			if tc.want != "" && !strings.Contains(got, tc.want) {
				t.Errorf("got %q; want substring %q", got, tc.want)
			}
		})
	}
}

func TestFormatURLPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://x.com/api/v2/foo?a=1", "/api/v2/foo"},
		{"https://x.com", "https://x.com"},      // no path -> return original
		{"plain-string", "plain-string"},          // no scheme/host -> return original
		{"", ""},
	}
	for _, tc := range cases {
		if got := FormatURLPath(tc.in); got != tc.want {
			t.Errorf("FormatURLPath(%q): got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatURLPath_NonString(t *testing.T) {
	got := FormatURLPath(42)
	if got != "42" {
		t.Errorf("non-string fallback: got %q", got)
	}
}

func TestFormatURLHost(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://region1.analytics.google.com/g/collect?x=1", "region1.analytics.google.com"},
		{"https://kmeqeuhj.1000farmacie.it/3hzxkmeqeuhj.js?x=1", "kmeqeuhj.1000farmacie.it"},
		{"plain", "plain"}, // no host -> return original
	}
	for _, tc := range cases {
		if got := FormatURLHost(tc.in); got != tc.want {
			t.Errorf("FormatURLHost(%q): got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	f := Truncate(5)
	cases := map[string]string{
		"short":         "short",
		"longerstring":  "long…",
		"":              "",
		"exactlyfive!!": "exac…",
	}
	for in, want := range cases {
		if got := f(in); got != want {
			t.Errorf("Truncate(5)(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestTruncate_NonString(t *testing.T) {
	f := Truncate(10)
	if got := f(42); got != "42" {
		t.Errorf("non-string Truncate: got %q", got)
	}
}

func TestTruncate_TinyN(t *testing.T) {
	f := Truncate(1)
	if got := f("xy"); got != "x" {
		t.Errorf("tiny N: got %q, want %q", got, "x")
	}
}

func TestToInt64(t *testing.T) {
	cases := []struct {
		in   any
		ok   bool
		want int64
	}{
		{int64(42), true, 42},
		{42, true, 42},
		{float64(42.7), true, 42},
		{"42", true, 42},
		{"abc", false, 0},
		{nil, false, 0},
		{[]int{1}, false, 0},
	}
	for _, tc := range cases {
		got, ok := toInt64(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("toInt64(%v): got (%d,%v), want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
