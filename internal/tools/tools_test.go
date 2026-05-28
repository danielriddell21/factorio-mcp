package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/danielriddell21/factorio-mcp/internal/factorio"
)

type fakeDispatcher struct {
	lastFn   string
	lastArgs any
	data     json.RawMessage
	err      error
}

func (f *fakeDispatcher) Call(_ context.Context, fn string, args any) (json.RawMessage, error) {
	f.lastFn = fn
	f.lastArgs = args
	return f.data, f.err
}

func (f *fakeDispatcher) Eval(_ context.Context, lua string) (json.RawMessage, error) {
	f.lastFn = "eval"
	f.lastArgs = lua
	return f.data, f.err
}

// listTools connects an in-memory client/server pair and returns the advertised
// tool names, exercising the real registration + schema generation path.
func listTools(t *testing.T, d factorio.Dispatcher, opts Options) map[string]bool {
	t.Helper()
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	Register(server, d, opts)

	ct, st := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	return names
}

func TestRegisterToolSurface(t *testing.T) {
	names := listTools(t, &fakeDispatcher{}, Options{})

	want := []string{
		"factorio_get_player_state", "factorio_get_inventory", "factorio_scan_entities",
		"factorio_scan_resources", "factorio_get_research_state", "factorio_get_tech_tree",
		"factorio_get_production_stats", "factorio_get_recipe_info",
		"factorio_place_entity", "factorio_remove_entity", "factorio_set_recipe",
		"factorio_insert_items", "factorio_remove_items", "factorio_craft",
		"factorio_set_research", "factorio_queue_research", "factorio_teleport",
		"factorio_mine_resource",
	}
	for _, n := range want {
		if !names[n] {
			t.Errorf("missing tool %q", n)
		}
	}
	if names["factorio_eval"] {
		t.Error("eval tool should be absent when AllowEval is false")
	}
}

func TestEvalGatedOn(t *testing.T) {
	names := listTools(t, &fakeDispatcher{}, Options{AllowEval: true})
	if !names["factorio_eval"] {
		t.Error("eval tool should be present when AllowEval is true")
	}
}

func TestToolCallDecodesAndForwards(t *testing.T) {
	ctx := context.Background()
	fd := &fakeDispatcher{data: json.RawMessage(`{"unit_number":42,"position":{"x":1,"y":2}}`)}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	Register(server, fd, Options{})

	ct, st := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "factorio_place_entity",
		Arguments: map[string]any{"name": "stone-furnace", "position": map[string]any{"x": 1, "y": 2}},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %+v", res.Content)
	}
	if fd.lastFn != "place_entity" {
		t.Fatalf("forwarded to wrong fn: %q", fd.lastFn)
	}

	// Structured output should reflect the dispatcher's data.
	b, _ := json.Marshal(res.StructuredContent)
	var out placeEntityOut
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode structured content: %v", err)
	}
	if out.UnitNumber != 42 {
		t.Fatalf("unit_number = %d, want 42", out.UnitNumber)
	}
}

func TestToolCallSurfacesError(t *testing.T) {
	ctx := context.Background()
	fd := &fakeDispatcher{err: &factorio.ErrLua{Code: "cant_place", Detail: "blocked"}}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	Register(server, fd, Options{})

	ct, st := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "factorio_place_entity",
		Arguments: map[string]any{"name": "x", "position": map[string]any{"x": 0, "y": 0}},
	})
	if err != nil {
		t.Fatalf("call tool transport error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError to be true")
	}
}
