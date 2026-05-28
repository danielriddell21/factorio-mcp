// Package config resolves runtime configuration for the Factorio MCP server
// from environment variables, with optional flag overrides handled by callers.
package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds everything needed to connect to a Factorio RCON endpoint.
type Config struct {
	// Host is the RCON host (default 127.0.0.1).
	Host string
	// Port is the RCON TCP port (default 27015).
	Port string
	// Password is the RCON password. Required.
	Password string
	// DialTimeout bounds the initial TCP connect + auth.
	DialTimeout time.Duration
	// RequestTimeout bounds a single command round-trip.
	RequestTimeout time.Duration
	// AllowEval enables the raw Lua eval escape hatch (a cheat / debug tool).
	AllowEval bool
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// Load builds a Config from environment variables and the provided flag set.
// Flags, when set, override the environment. It returns an error if the
// password is missing, since the server cannot connect without it.
func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("factorio-mcp", flag.ContinueOnError)

	host := fs.String("host", env("FACTORIO_RCON_HOST", "127.0.0.1"), "Factorio RCON host")
	port := fs.String("port", env("FACTORIO_RCON_PORT", "27015"), "Factorio RCON port")
	password := fs.String("password", env("FACTORIO_RCON_PASSWORD", ""), "Factorio RCON password")
	allowEval := fs.Bool("allow-eval", env("FACTORIO_MCP_ALLOW_EVAL", "") != "", "enable the raw Lua eval tool (cheat/debug)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if *password == "" {
		return Config{}, fmt.Errorf("RCON password is required: set FACTORIO_RCON_PASSWORD or pass -password")
	}

	return Config{
		Host:           *host,
		Port:           *port,
		Password:       *password,
		DialTimeout:    10 * time.Second,
		RequestTimeout: 15 * time.Second,
		AllowEval:      *allowEval,
	}, nil
}

// Address returns the host:port string used to dial RCON.
func (c Config) Address() string {
	return c.Host + ":" + c.Port
}
