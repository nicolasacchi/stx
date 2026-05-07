package timeparse

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	now := time.Now().Unix()

	cases := []struct {
		name    string
		in      string
		wantErr bool
		check   func(int64) bool
	}{
		{"empty -> now", "", false, func(v int64) bool { return abs(v-now) <= 2 }},
		{"now", "now", false, func(v int64) bool { return abs(v-now) <= 2 }},
		{"1h", "1h", false, func(v int64) bool { return abs((now-3600)-v) <= 2 }},
		{"7d", "7d", false, func(v int64) bool { return abs((now-7*86400)-v) <= 2 }},
		{"now-2h", "now-2h", false, func(v int64) bool { return abs((now-7200)-v) <= 2 }},
		{"epoch seconds", "1769681520", false, func(v int64) bool { return v == 1769681520 }},
		{"epoch millis", "1769681520000", false, func(v int64) bool { return v == 1769681520 }},
		{"RFC3339", "2026-03-19T09:00:00Z", false, func(v int64) bool {
			t, _ := time.Parse(time.RFC3339, "2026-03-19T09:00:00Z")
			return v == t.Unix()
		}},
		{"garbage", "not-a-time", true, nil},
		{"unknown unit", "1y", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.check(got) {
				t.Fatalf("check failed: got %d", got)
			}
		})
	}
}

func TestParseMillis(t *testing.T) {
	got, err := ParseMillis("1h")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	if abs(got-(now-3600*1000)) > 2000 {
		t.Fatalf("expected ~1h ago in millis; got %d (now=%d)", got, now)
	}
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
