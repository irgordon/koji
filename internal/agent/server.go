package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"koji/internal/config"
)

type Server struct {
	socketPath string
	listener   net.Listener
	executor   Executor
}

func NewServer(socketPath string) *Server {
	return NewServerWithConfig(config.Config{AgentSocketPath: socketPath})
}

func NewServerWithConfig(cfg config.Config) *Server {
	return &Server{
		socketPath: cfg.AgentSocketPath,
		executor:   NewExecutor(cfg),
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := listenUnixSocket(s.socketPath)
	if err != nil {
		return err
	}
	s.listener = listener
	defer s.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.serve()
	}()

	select {
	case <-ctx.Done():
		_ = s.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Close() error {
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	_ = removeSocketFile(s.socketPath)
	return err
}

func (s *Server) serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept agent connection: %w", err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	var request RPCRequest
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		writeRPCResponse(conn, RPCResponse{Status: ResponseError, Reason: "malformed_request"})
		return
	}

	writeRPCResponse(conn, s.responseForRequest(request))
}

func (s *Server) responseForRequest(request RPCRequest) RPCResponse {
	if request.Method != MethodServiceControl {
		return RPCResponse{Status: ResponseError, Reason: "unknown_method"}
	}
	return s.executor.ControlService(context.Background(), request.Service)
}

func writeRPCResponse(conn net.Conn, response RPCResponse) {
	_ = json.NewEncoder(conn).Encode(response)
}

func listenUnixSocket(socketPath string) (net.Listener, error) {
	if err := ValidateSocketPath(socketPath); err != nil {
		return nil, err
	}
	if err := validateSocketParent(socketPath); err != nil {
		return nil, err
	}
	if err := removeStaleSocket(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("listen agent socket: %w", err)
	}
	return listener, nil
}

func validateSocketParent(socketPath string) error {
	parent := filepath.Dir(socketPath)
	info, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("stat agent socket parent: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("agent socket parent is not a directory")
	}
	if isUnsafeWorldWritableParent(info) {
		return fmt.Errorf("agent socket parent is world-writable without sticky root ownership")
	}
	return nil
}

func isUnsafeWorldWritableParent(info os.FileInfo) bool {
	if info.Mode().Perm()&0002 == 0 {
		return false
	}
	return !hasStickyRootOwnedDirectory(info)
}

func removeStaleSocket(socketPath string) error {
	info, err := os.Lstat(socketPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat existing agent socket: %w", err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("refuse to remove non-socket agent path")
	}
	if err := os.Remove(socketPath); err != nil {
		return fmt.Errorf("remove stale agent socket: %w", err)
	}
	return nil
}

func removeSocketFile(socketPath string) error {
	info, err := os.Lstat(socketPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSocket == 0 {
		return nil
	}
	return os.Remove(socketPath)
}
