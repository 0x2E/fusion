package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func legacyString(value any, fallback string) string {
	if value == nil {
		return fallback
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func legacyInt64(value any, fallback int64) int64 {
	v, ok := parseLegacyInt64(value)
	if !ok {
		return fallback
	}
	return v
}

func legacyUnixSeconds(value any, fallback int64) int64 {
	v, ok := parseLegacyUnixSeconds(value)
	if !ok {
		return fallback
	}
	return v
}

func legacyUnixSecondsNullable(value any) (int64, bool) {
	return parseLegacyUnixSeconds(value)
}

func parseLegacyInt64(value any) (int64, bool) {
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case bool:
		return boolToInt64(v), true
	case []byte:
		return parseIntString(string(v))
	case string:
		return parseIntString(v)
	case time.Time:
		return v.Unix(), true
	default:
		return 0, false
	}
}

func parseLegacyUnixSeconds(value any) (int64, bool) {
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case time.Time:
		return v.Unix(), true
	case string:
		return parseUnixString(v)
	case []byte:
		return parseUnixString(string(v))
	default:
		return parseLegacyInt64(value)
	}
}

func parseIntString(raw string) (int64, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, false
	}

	if v, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return v, true
	}

	if v, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return int64(v), true
	}

	return 0, false
}

func parseUnixString(raw string) (int64, bool) {
	if v, ok := parseIntString(raw); ok {
		return v, true
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if ts, ok := parseWithLayout(layout, trimmed); ok {
			return ts, true
		}
	}

	return 0, false
}

func parseWithLayout(layout, value string) (int64, bool) {
	if t, err := time.Parse(layout, value); err == nil {
		return t.Unix(), true
	}
	if t, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
		return t.Unix(), true
	}
	return 0, false
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
