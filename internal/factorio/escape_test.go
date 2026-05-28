package factorio

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEscapeLuaSingleQuoted(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", "hello"},
		{"json double quotes pass through", `{"a":1}`, `{"a":1}`},
		{"single quote", "it's", `it\'s`},
		{"backslash", `a\b`, `a\\b`},
		{"backslash before quote", `\'`, `\\\'`},
		{"newline", "a\nb", `a\nb`},
		{"carriage return", "a\rb", `a\rb`},
		{"tab", "a\tb", `a\tb`},
		{"nul", "a\x00b", `a\0b`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EscapeLuaSingleQuoted(tc.in); got != tc.want {
				t.Fatalf("EscapeLuaSingleQuoted(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// unquoteLuaSingle reverses the subset of Lua single-quoted string escaping
// that EscapeLuaSingleQuoted produces. It lets the fuzz test assert a clean
// round-trip without embedding a full Lua interpreter.
func unquoteLuaSingle(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case '\'':
			b.WriteByte('\'')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case '0':
			b.WriteByte(0)
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func FuzzEscapeRoundTrip(f *testing.F) {
	seeds := []any{
		map[string]any{"fn": "place_entity", "args": map[string]any{"name": "stone-furnace's", "pos": []float64{1.5, -2.25}}},
		map[string]any{"weird": "back\\slash and \"quotes\" and 'apostrophes'"},
		map[string]any{"unicode": "résumé — 日本語"},
	}
	for _, s := range seeds {
		b, _ := json.Marshal(s)
		f.Add(string(b))
	}
	f.Fuzz(func(t *testing.T, payload string) {
		escaped := EscapeLuaSingleQuoted(payload)
		if got := unquoteLuaSingle(escaped); got != payload {
			t.Fatalf("round-trip failed: payload=%q escaped=%q got=%q", payload, escaped, got)
		}
		// The escaped form must never contain a bare single quote, which would
		// prematurely close the Lua string literal.
		for i := 0; i < len(escaped); i++ {
			if escaped[i] == '\'' && (i == 0 || escaped[i-1] != '\\') {
				t.Fatalf("escaped form contains a bare single quote at %d: %q", i, escaped)
			}
		}
	})
}
