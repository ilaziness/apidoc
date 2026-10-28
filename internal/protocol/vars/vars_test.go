package vars

import "testing"

func TestExpand(t *testing.T) {
	m := map[string]string{"baseUrl": "https://api.example.com", "token": "abc"}
	cases := []struct {
		in, want string
	}{
		{"{{baseUrl}}/v1", "https://api.example.com/v1"},
		{"Bearer {{token}}", "Bearer abc"},
		{"{{missing}}", "{{missing}}"},
		{"{{ baseUrl }}", "{{ baseUrl }}"}, // no trim
		{"{{BaseUrl}}", "{{BaseUrl}}"},     // case sensitive
		{"no vars", "no vars"},
		{"{{baseUrl}}{{token}}", "https://api.example.comabc"},
		{"{{nested{{baseUrl}}}}", "{{nested{{baseUrl}}}}"}, // no nesting support leaves outer intact when key wrong
	}
	for _, c := range cases {
		if got := Expand(c.in, m); got != c.want {
			t.Errorf("Expand(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
