package output

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// FormatFunc transforms a raw value into a display string.
type FormatFunc func(value any) string

// FormatEpochSeconds — render integer seconds-since-epoch as local RFC3339.
func FormatEpochSeconds(v any) string {
	n, ok := toInt64(v)
	if !ok || n == 0 {
		return ""
	}
	return time.Unix(n, 0).Local().Format("2006-01-02 15:04:05")
}

// FormatURLPath — strip scheme/host/query, keep just the path.
func FormatURLPath(v any) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	if u.Path == "" {
		return s
	}
	return u.Path
}

// FormatURLHost — extract just the hostname.
func FormatURLHost(v any) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return s
	}
	return u.Host
}

// Truncate returns a FormatFunc that caps strings at n chars (with ellipsis).
func Truncate(n int) FormatFunc {
	return func(v any) string {
		s := fmt.Sprintf("%v", v)
		if len(s) <= n {
			return s
		}
		if n <= 1 {
			return s[:n]
		}
		return s[:n-1] + "…"
	}
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}
