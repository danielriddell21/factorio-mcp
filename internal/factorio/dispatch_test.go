package factorio

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeExec struct {
	lastCmd string
	out     string
	err     error
}

func (f *fakeExec) Execute(_ context.Context, command string) (string, error) {
	f.lastCmd = command
	return f.out, f.err
}

func TestCallSuccess(t *testing.T) {
	fe := &fakeExec{out: `{"ok":true,"data":{"x":1.5,"y":2}}`}
	c := NewClient(fe)

	data, err := c.Call(context.Background(), "get_player_state", map[string]any{"player_index": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"x":1.5,"y":2}` {
		t.Fatalf("unexpected data: %s", data)
	}

	// The command must be a single-line silent-command carrying the payload.
	if !strings.HasPrefix(fe.lastCmd, "/silent-command ") {
		t.Fatalf("command not a silent-command: %q", fe.lastCmd)
	}
	if strings.Contains(fe.lastCmd, "\n") {
		t.Fatalf("command must be a single line: %q", fe.lastCmd)
	}
	// The payload sits inside a Lua single-quoted literal, so its JSON double
	// quotes are not escaped.
	if !strings.Contains(fe.lastCmd, `'{"fn":"get_player_state"`) {
		t.Fatalf("payload missing fn: %q", fe.lastCmd)
	}
}

func TestCallLuaError(t *testing.T) {
	fe := &fakeExec{out: `{"ok":false,"error":"cant_place","detail":"colliding with iron-ore"}`}
	c := NewClient(fe)

	_, err := c.Call(context.Background(), "place_entity", nil)
	var le *ErrLua
	if !errors.As(err, &le) {
		t.Fatalf("expected *ErrLua, got %T: %v", err, err)
	}
	if le.Code != "cant_place" || le.Detail != "colliding with iron-ore" {
		t.Fatalf("unexpected ErrLua: %+v", le)
	}
	if !strings.Contains(le.Error(), "colliding with iron-ore") {
		t.Fatalf("error message should surface detail: %q", le.Error())
	}
}

func TestCallRawNonJSON(t *testing.T) {
	fe := &fakeExec{out: "Error: unknown command"}
	c := NewClient(fe)

	_, err := c.Call(context.Background(), "whatever", nil)
	var le *ErrLua
	if !errors.As(err, &le) {
		t.Fatalf("expected *ErrLua, got %T: %v", err, err)
	}
	if le.Raw != "Error: unknown command" {
		t.Fatalf("raw not preserved: %q", le.Raw)
	}
}

func TestCallExecError(t *testing.T) {
	sentinel := errors.New("boom")
	fe := &fakeExec{err: sentinel}
	c := NewClient(fe)

	_, err := c.Call(context.Background(), "x", nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected exec error to propagate, got %v", err)
	}
}
