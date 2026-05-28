// Package rcon provides a Factorio-aware wrapper around a Source RCON
// connection. It serializes access, connects lazily, and transparently
// reconnects+retries once when the connection drops.
package rcon

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gorcon/rcon"

	"github.com/danielriddell21/factorio-mcp/internal/config"
)

// ErrConnection indicates the RCON endpoint could not be reached or the
// command could not be delivered after a reconnect attempt.
type ErrConnection struct{ Err error }

func (e *ErrConnection) Error() string { return fmt.Sprintf("rcon connection error: %v", e.Err) }
func (e *ErrConnection) Unwrap() error { return e.Err }

// Executor is the minimal surface the rest of the program depends on, which
// makes it trivial to substitute a fake in tests.
type Executor interface {
	Execute(ctx context.Context, command string) (string, error)
}

// Client is a serialized, self-healing RCON client.
type Client struct {
	cfg config.Config

	mu   sync.Mutex
	conn *rcon.Conn
}

// New constructs a Client. It does not connect; the first Execute does.
func New(cfg config.Config) *Client {
	return &Client{cfg: cfg}
}

// Close releases the underlying connection, if any.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) dial() error {
	conn, err := rcon.Dial(
		c.cfg.Address(),
		c.cfg.Password,
		rcon.SetDialTimeout(c.cfg.DialTimeout),
		rcon.SetDeadline(c.cfg.RequestTimeout),
	)
	if err != nil {
		return &ErrConnection{Err: err}
	}
	c.conn = conn
	return nil
}

func (c *Client) drop() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// Execute sends a single command and returns the response body. It connects
// lazily and, if the live connection has dropped, reconnects once and retries
// the command a single time before surfacing an *ErrConnection.
func (c *Client) Execute(ctx context.Context, command string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Up to two attempts: the first may fail on a stale connection, the second
	// runs against a freshly dialed one.
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if c.conn == nil {
			if err := c.dial(); err != nil {
				lastErr = err
				continue
			}
		}

		out, err := c.conn.Execute(command)
		if err == nil {
			return out, nil
		}

		// A failed Execute usually means the socket is unusable; drop it so the
		// next attempt redials.
		lastErr = err
		c.drop()
	}

	if lastErr == nil {
		lastErr = errors.New("unknown rcon failure")
	}
	return "", &ErrConnection{Err: lastErr}
}
