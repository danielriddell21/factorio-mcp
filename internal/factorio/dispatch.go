// Package factorio turns typed calls into one-line RCON commands routed through
// the factorio_mcp companion mod, and decodes the mod's JSON envelope back into
// Go values.
package factorio

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/danielriddell21/factorio-mcp/internal/rcon"
)

// ErrLua reports an error returned by the companion mod (a handler error, a bad
// request, or non-JSON output from the game engine).
type ErrLua struct {
	Code   string
	Detail string
	Raw    string
}

func (e *ErrLua) Error() string {
	switch {
	case e.Code != "" && e.Detail != "":
		return fmt.Sprintf("factorio error %q: %s", e.Code, e.Detail)
	case e.Code != "":
		return fmt.Sprintf("factorio error: %s", e.Code)
	default:
		return fmt.Sprintf("factorio error: %s", strings.TrimSpace(e.Raw))
	}
}

// envelope is the JSON contract returned by the mod's dispatch function.
type envelope struct {
	OK     bool            `json:"ok"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
	Detail string          `json:"detail"`
}

// request is the payload handed to the mod: a function name plus its arguments.
type request struct {
	Fn   string `json:"fn"`
	Args any    `json:"args"`
}

// Dispatcher executes named companion-mod functions and evaluates raw Lua.
type Dispatcher interface {
	Call(ctx context.Context, fn string, args any) (json.RawMessage, error)
	Eval(ctx context.Context, lua string) (json.RawMessage, error)
}

// Client is the default Dispatcher, backed by an RCON executor.
type Client struct {
	exec rcon.Executor
}

// NewClient wraps an RCON executor in a Dispatcher.
func NewClient(exec rcon.Executor) *Client {
	return &Client{exec: exec}
}

// Call invokes the named mod function with the given arguments and returns the
// decoded "data" field of the response envelope.
func (c *Client) Call(ctx context.Context, fn string, args any) (json.RawMessage, error) {
	if args == nil {
		args = struct{}{}
	}
	payload, err := json.Marshal(request{Fn: fn, Args: args})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	cmd := renderDispatch(EscapeLuaSingleQuoted(string(payload)))
	return c.run(ctx, cmd)
}

// Eval runs raw Lua source through the mod's eval handler. It is a cheat/debug
// escape hatch; callers gate it behind configuration.
func (c *Client) Eval(ctx context.Context, lua string) (json.RawMessage, error) {
	cmd := renderEval(EscapeLuaSingleQuoted(lua))
	return c.run(ctx, cmd)
}

func (c *Client) run(ctx context.Context, cmd string) (json.RawMessage, error) {
	out, err := c.exec.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return nil, &ErrLua{Code: "empty_response", Detail: "the game returned no output"}
	}

	var env envelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		// The engine prints raw Lua/console errors as plain text on failure.
		return nil, &ErrLua{Raw: out}
	}
	if !env.OK {
		return nil, &ErrLua{Code: env.Error, Detail: env.Detail}
	}
	return env.Data, nil
}

// Compile-time assertion that Client satisfies Dispatcher.
var _ Dispatcher = (*Client)(nil)
