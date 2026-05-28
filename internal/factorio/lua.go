package factorio

import (
	_ "embed"
	"strings"
)

// payloadPlaceholder is the token replaced with the (escaped) payload in the
// embedded Lua command templates.
const payloadPlaceholder = "__PAYLOAD__"

//go:embed templates/dispatch.lua
var dispatchTemplate string

//go:embed templates/eval.lua
var evalTemplate string

// EscapeLuaSingleQuoted escapes s so it can be safely embedded inside a Lua
// single-quoted string literal. Payloads are always passed inside '...' in the
// command templates, so only a small, well-defined set of bytes is dangerous:
// the escape character itself and the closing quote. Control characters are
// escaped defensively even though JSON-encoded payloads never contain raw ones.
func EscapeLuaSingleQuoted(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case 0:
			b.WriteString(`\0`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// renderDispatch builds the one-line RCON command for an escaped JSON payload.
func renderDispatch(escapedPayload string) string {
	return strings.TrimRight(strings.Replace(dispatchTemplate, payloadPlaceholder, escapedPayload, 1), "\n")
}

// renderEval builds the one-line RCON command for an escaped Lua source payload.
func renderEval(escapedPayload string) string {
	return strings.TrimRight(strings.Replace(evalTemplate, payloadPlaceholder, escapedPayload, 1), "\n")
}
