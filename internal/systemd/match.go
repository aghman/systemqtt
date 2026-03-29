package systemd

import (
	"path/filepath"
	"strings"
)

// MatchUnit returns true if name matches any non-empty glob in patterns, or if patterns is empty (all units).
func MatchUnit(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		ok, err := filepath.Match(p, name)
		if err == nil && ok {
			return true
		}
	}
	return false
}

// SplitPatterns splits a comma-separated filter string.
func SplitPatterns(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
