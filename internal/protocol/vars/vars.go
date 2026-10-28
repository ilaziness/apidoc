// Package vars replaces {{key}} placeholders per PRD rules:
// case-sensitive, no trim, no nesting; unmatched tokens stay as-is.
package vars

import (
	"strings"
)

// Expand replaces {{key}} occurrences using vars. Unmatched placeholders remain.
func Expand(s string, vars map[string]string) string {
	if s == "" || len(vars) == 0 && !strings.Contains(s, "{{") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		start := strings.Index(s[i:], "{{")
		if start < 0 {
			b.WriteString(s[i:])
			break
		}
		start += i
		b.WriteString(s[i:start])
		end := strings.Index(s[start+2:], "}}")
		if end < 0 {
			b.WriteString(s[start:])
			break
		}
		end = start + 2 + end
		key := s[start+2 : end]
		if val, ok := vars[key]; ok {
			b.WriteString(val)
		} else {
			b.WriteString(s[start : end+2])
		}
		i = end + 2
	}
	return b.String()
}
