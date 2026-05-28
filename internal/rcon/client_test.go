package rcon

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielriddell21/factorio-mcp/internal/config"
)

// Source RCON packet types.
const (
	typeAuth         = 3
	typeAuthResponse = 2
	typeExecCommand  = 2
	typeResponse     = 0
)

func readPacket(conn net.Conn) (id, typ int32, body string, err error) {
	var size int32
	if err = binary.Read(conn, binary.LittleEndian, &size); err != nil {
		return
	}
	buf := make([]byte, size)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return
	}
	id = int32(binary.LittleEndian.Uint32(buf[0:4]))
	typ = int32(binary.LittleEndian.Uint32(buf[4:8]))
	body = string(buf[8 : len(buf)-2]) // strip two trailing NULs
	return
}

func writePacket(conn net.Conn, id, typ int32, body string) error {
	payload := make([]byte, 8+len(body)+2)
	binary.LittleEndian.PutUint32(payload[0:4], uint32(id))
	binary.LittleEndian.PutUint32(payload[4:8], uint32(typ))
	copy(payload[8:], body)
	out := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint32(out[0:4], uint32(len(payload)))
	copy(out[4:], payload)
	_, err := conn.Write(out)
	return err
}

// mockServer is a minimal Source RCON server for tests.
type mockServer struct {
	ln        net.Listener
	password  string
	respond   func(cmd string) string // command -> response body
	closeNext atomic.Bool             // drop the next accepted connection right after auth
}

func newMockServer(t *testing.T, password string, respond func(string) string) *mockServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &mockServer{ln: ln, password: password, respond: respond}
	go s.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *mockServer) addr() string { return s.ln.Addr().String() }

func (s *mockServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *mockServer) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	id, typ, body, err := readPacket(conn)
	if err != nil || typ != typeAuth {
		return
	}
	if body != s.password {
		_ = writePacket(conn, -1, typeAuthResponse, "") // auth failure id = -1
		return
	}
	if err := writePacket(conn, id, typeAuthResponse, ""); err != nil {
		return
	}

	if s.closeNext.Swap(false) {
		return // simulate a dropped connection after a successful auth
	}

	for {
		id, typ, body, err := readPacket(conn)
		if err != nil {
			return
		}
		if typ != typeExecCommand {
			continue
		}
		if err := writePacket(conn, id, typeResponse, s.respond(body)); err != nil {
			return
		}
	}
}

func testConfig(addr, password string) config.Config {
	host, port, _ := net.SplitHostPort(addr)
	return config.Config{
		Host:           host,
		Port:           port,
		Password:       password,
		DialTimeout:    2 * time.Second,
		RequestTimeout: 2 * time.Second,
	}
}

func TestExecuteSuccess(t *testing.T) {
	srv := newMockServer(t, "secret", func(cmd string) string {
		if cmd == "ping" {
			return "pong"
		}
		return "unknown"
	})
	c := New(testConfig(srv.addr(), "secret"))
	defer func() { _ = c.Close() }()

	out, err := c.Execute(context.Background(), "ping")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out != "pong" {
		t.Fatalf("got %q, want pong", out)
	}
}

func TestExecuteAuthFailure(t *testing.T) {
	srv := newMockServer(t, "secret", func(string) string { return "" })
	c := New(testConfig(srv.addr(), "wrong"))
	defer func() { _ = c.Close() }()

	_, err := c.Execute(context.Background(), "ping")
	var ce *ErrConnection
	if !errors.As(err, &ce) {
		t.Fatalf("expected *ErrConnection on auth failure, got %T: %v", err, err)
	}
}

func TestExecuteReconnectAndRetry(t *testing.T) {
	srv := newMockServer(t, "secret", func(string) string { return "ok" })
	c := New(testConfig(srv.addr(), "secret"))
	defer func() { _ = c.Close() }()

	// Establish a live connection.
	if _, err := c.Execute(context.Background(), "warmup"); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	// Force the next server-side connection to drop right after auth, and kill
	// the current client connection so the next Execute must redial.
	srv.closeNext.Store(true)
	c.drop()

	// First redial hits the dropped connection; the client should redial again
	// and succeed within its two attempts.
	out, err := c.Execute(context.Background(), "again")
	if err != nil {
		t.Fatalf("expected success after reconnect, got %v", err)
	}
	if out != "ok" {
		t.Fatalf("got %q, want ok", out)
	}
}

func TestExecuteContextCanceled(t *testing.T) {
	srv := newMockServer(t, "secret", func(string) string { return "ok" })
	c := New(testConfig(srv.addr(), "secret"))
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.Execute(ctx, "ping"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestExecuteServerDown(t *testing.T) {
	// Listen then immediately close to get a guaranteed-dead address.
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()

	c := New(testConfig(addr, "secret"))
	_, err := c.Execute(context.Background(), "ping")
	var ce *ErrConnection
	if !errors.As(err, &ce) {
		t.Fatalf("expected *ErrConnection, got %T: %v", err, err)
	}
}
