package ui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
)

func idString(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

func parseID(vals []string) (int64, bool) {
	if len(vals) == 0 || vals[0] == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(vals[0], 10, 64)
	return id, err == nil
}

func selectedOrEmpty(id string) []string {
	if id == "" {
		return nil
	}
	return []string{id}
}

func encodeKV(rows []httpapi.KV) string {
	var b strings.Builder
	for _, r := range rows {
		if !r.Enabled {
			b.WriteByte('-')
		}
		b.WriteString(r.Key)
		b.WriteByte(':')
		b.WriteString(r.Value)
		b.WriteByte('\n')
	}
	return b.String()
}

func decodeKV(s string) []httpapi.KV {
	lines := strings.Split(s, "\n")
	out := make([]httpapi.KV, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		enabled := true
		if strings.HasPrefix(line, "-") {
			enabled = false
			line = line[1:]
		}
		key, val, _ := strings.Cut(line, ":")
		out = append(out, httpapi.KV{Key: key, Value: val, Enabled: enabled})
	}
	return out
}

func encodeMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(m[k])
		b.WriteByte('\n')
	}
	return b.String()
}

func decodeMap(s string) map[string]string {
	out := map[string]string{}
	for _, kv := range decodeKV(s) {
		if kv.Key == "" {
			continue
		}
		out[kv.Key] = kv.Value
	}
	return out
}
