package agent

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestUnixSocketClientServerReturnsMutationDisabled(t *testing.T) {
	socketPath := shortSocketPath(t)
	server := NewServer(socketPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServer(t, ctx, server, socketPath)

	err := NewClient(socketPath).ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})
	if !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("expected mutation disabled, got %v", err)
	}
}

func TestClientReturnsUnavailableForMissingSocket(t *testing.T) {
	socketPath := shortSocketPath(t)

	err := NewClient(socketPath).ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})
	if !errors.Is(err, ErrAgentUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestClassifyDialErrorReturnsConnectionRefused(t *testing.T) {
	err := classifyDialError(context.Background(), syscall.ECONNREFUSED)
	if !errors.Is(err, ErrConnectionRefused) {
		t.Fatalf("expected connection refused, got %v", err)
	}
}

func TestClientReturnsTimeout(t *testing.T) {
	socketPath := shortSocketPath(t)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		skipIfUnixSocketBindDenied(t, err)
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()
	go acceptAndHold(t, listener)

	client := NewClient(socketPath)
	client.timeout = 20 * time.Millisecond

	err = client.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})
	if !errors.Is(err, ErrAgentTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestClientReturnsMalformedResponse(t *testing.T) {
	socketPath := shortSocketPath(t)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		skipIfUnixSocketBindDenied(t, err)
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()
	go acceptMalformedResponse(t, listener)

	err = NewClient(socketPath).ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})
	if !errors.Is(err, ErrMalformedResponse) {
		t.Fatalf("expected malformed response, got %v", err)
	}
}

func TestServerRejectsRelativeSocketPath(t *testing.T) {
	err := listenServerOnce("relative.sock")
	if !errors.Is(err, ErrInvalidSocketPath) {
		t.Fatalf("expected invalid socket path, got %v", err)
	}
}

func TestServerRejectsUnsafeWorldWritableParent(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "unsafe")
	if err := os.Mkdir(parent, 0777); err != nil {
		t.Fatalf("create unsafe parent: %v", err)
	}
	if err := os.Chmod(parent, 0777); err != nil {
		t.Fatalf("chmod unsafe parent: %v", err)
	}

	err := listenServerOnce(filepath.Join(parent, "agent.sock"))
	if err == nil {
		t.Fatal("expected unsafe parent error")
	}
}

func TestServerRefusesToRemoveNonSocketPath(t *testing.T) {
	socketPath := shortSocketPath(t)
	if err := os.WriteFile(socketPath, []byte("not a socket"), 0600); err != nil {
		t.Fatalf("write non-socket file: %v", err)
	}

	err := listenServerOnce(socketPath)
	if err == nil {
		t.Fatal("expected non-socket path error")
	}
}

func shortSocketPath(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("/tmp", "koji-agent-*")
	if err != nil {
		t.Fatalf("create short socket dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return filepath.Join(dir, "a.sock")
}

func startServer(t *testing.T, ctx context.Context, server *Server, socketPath string) {
	t.Helper()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	for i := 0; i < 100; i++ {
		if _, err := os.Stat(socketPath); err == nil {
			return
		}
		select {
		case err := <-errCh:
			skipIfUnixSocketBindDenied(t, err)
			t.Fatalf("server exited early: %v", err)
		default:
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("server did not create socket")
}

func skipIfUnixSocketBindDenied(t *testing.T, err error) {
	t.Helper()

	if os.IsPermission(err) || errors.Is(err, syscall.EPERM) {
		t.Skipf("unix socket bind not permitted in this environment: %v", err)
	}
}

func listenServerOnce(socketPath string) error {
	listener, err := listenUnixSocket(socketPath)
	if listener != nil {
		_ = listener.Close()
	}
	return err
}

func acceptAndHold(t *testing.T, listener net.Listener) {
	t.Helper()

	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	time.Sleep(100 * time.Millisecond)
}

func acceptMalformedResponse(t *testing.T, listener net.Listener) {
	t.Helper()

	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("not-json\n"))
}
