// Package tools registers the MCP tool surface that Claude uses to perceive and
// act on a running Factorio game. Each tool maps a typed input directly to a
// named companion-mod function and decodes the typed result.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/danielriddell21/factorio-mcp/internal/factorio"
)

// Options controls which tools are registered.
type Options struct {
	// AllowEval registers the raw Lua eval escape hatch (a cheat/debug tool).
	AllowEval bool
}

// Register adds every Factorio tool to the server.
func Register(s *mcp.Server, d factorio.Dispatcher, opts Options) {
	registerIntrospection(s, d)
	registerActions(s, d)
	if opts.AllowEval {
		registerEval(s, d)
	}
}

// addTool wires a typed input/output pair to a named mod function. Returning an
// error from the handler is surfaced to the model as a tool error whose message
// includes the mod's own error string, so Claude can read it and self-correct.
func addTool[In, Out any](s *mcp.Server, d factorio.Dispatcher, name, description, fn string) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: description},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
			var out Out
			data, err := d.Call(ctx, fn, in)
			if err != nil {
				return nil, out, err
			}
			if len(data) > 0 && string(data) != "null" {
				if err := json.Unmarshal(data, &out); err != nil {
					return nil, out, fmt.Errorf("decode %s response: %w", fn, err)
				}
			}
			return nil, out, nil
		})
}
