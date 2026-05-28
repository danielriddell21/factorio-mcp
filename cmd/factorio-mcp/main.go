// Command factorio-mcp is an MCP server that lets Claude play Factorio by
// reading game state and issuing high-level actions over RCON, routed through
// the factorio_mcp companion mod.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/danielriddell21/factorio-mcp/internal/config"
	"github.com/danielriddell21/factorio-mcp/internal/factorio"
	"github.com/danielriddell21/factorio-mcp/internal/rcon"
	"github.com/danielriddell21/factorio-mcp/internal/tools"
)

// version is overridable at build time via -ldflags.
var version = "0.1.0"

func main() {
	if err := run(); err != nil {
		log.Fatalf("factorio-mcp: %v", err)
	}
}

func run() error {
	// stdout is reserved for the MCP stdio channel; all logs go to stderr.
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := rcon.New(cfg)
	defer func() { _ = client.Close() }()

	dispatcher := factorio.NewClient(client)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "factorio-mcp",
		Title:   "Factorio MCP",
		Version: version,
	}, nil)

	tools.Register(server, dispatcher, tools.Options{AllowEval: cfg.AllowEval})

	log.Printf("factorio-mcp %s: RCON %s, eval=%v", version, cfg.Address(), cfg.AllowEval)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}
