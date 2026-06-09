package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"koji/internal/system"
)

var (
	ErrAgentUnavailable         = errors.New("agent service control unavailable")
	ErrConnectionRefused        = errors.New("agent connection refused")
	ErrAgentTimeout             = errors.New("agent request timeout")
	ErrMalformedResponse        = errors.New("agent malformed response")
	ErrNotImplemented           = errors.New("agent operation not implemented")
	ErrMutationDisabled         = errors.New("agent mutation disabled")
	ErrServiceNotAllowlisted    = errors.New("agent service not allowlisted")
	ErrCommandFailed            = errors.New("agent command failed")
	ErrInvalidSocketPath        = errors.New("invalid agent socket path")
	ErrInvalidServiceName       = errors.New("invalid service name")
	ErrUnsupportedServiceAction = errors.New("unsupported service action")
)

const (
	MethodServiceControl        = "service_control"
	ResponseOK                  = "ok"
	ResponseError               = "error"
	ReasonNotImplemented        = "not_implemented"
	ReasonMutationDisabled      = "mutation_disabled"
	ReasonServiceNotAllowlisted = "service_not_allowlisted"
	ReasonUnsupportedAction     = "unsupported_action"
	ReasonValidationError       = "validation_error"
	ReasonCommandFailed         = "command_failed"
	ReasonCommandTimeout        = "command_timeout"
	defaultTimeout              = 2 * time.Second
)

type ServiceControlRequest struct {
	Service string
	Action  string
}

type RPCRequest struct {
	Method  string                `json:"method"`
	Service ServiceControlRequest `json:"service,omitempty"`
}

type RPCResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type ServiceController interface {
	ControlService(ctx context.Context, request ServiceControlRequest) error
}

type Client struct {
	socketPath string
	timeout    time.Duration
}

func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    defaultTimeout,
	}
}

func (c *Client) ControlService(ctx context.Context, request ServiceControlRequest) error {
	if err := ValidateServiceControlRequest(request); err != nil {
		return err
	}
	if err := ValidateSocketPath(c.socketPath); err != nil {
		return err
	}

	response, err := c.call(ctx, RPCRequest{
		Method:  MethodServiceControl,
		Service: request,
	})
	if err != nil {
		return err
	}

	return responseError(response)
}

func CheckReachable(ctx context.Context, socketPath string) error {
	client := NewClient(socketPath)
	if err := ValidateSocketPath(client.socketPath); err != nil {
		return err
	}
	conn, err := client.dial(ctx)
	if err != nil {
		return err
	}
	return conn.Close()
}

func (c *Client) call(ctx context.Context, request RPCRequest) (RPCResponse, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return RPCResponse{}, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return RPCResponse{}, fmt.Errorf("%w: write request", ErrAgentUnavailable)
	}

	var response RPCResponse
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&response); err != nil {
		if isTimeout(err) {
			return RPCResponse{}, ErrAgentTimeout
		}
		return RPCResponse{}, ErrMalformedResponse
	}
	if response.Status == "" {
		return RPCResponse{}, ErrMalformedResponse
	}
	return response, nil
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "unix", c.socketPath)
	if err == nil {
		return conn, nil
	}
	return nil, classifyDialError(dialCtx, err)
}

func classifyDialError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || isTimeout(err) {
		return ErrAgentTimeout
	}
	if errors.Is(err, os.ErrNotExist) {
		return ErrAgentUnavailable
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return ErrConnectionRefused
	}
	return fmt.Errorf("%w: %v", ErrAgentUnavailable, err)
}

func responseError(response RPCResponse) error {
	if response.Status == ResponseOK {
		return nil
	}
	if response.Status != ResponseError {
		return ErrMalformedResponse
	}
	switch response.Reason {
	case ReasonNotImplemented:
		return ErrNotImplemented
	case ReasonMutationDisabled:
		return ErrMutationDisabled
	case ReasonServiceNotAllowlisted:
		return ErrServiceNotAllowlisted
	case ReasonUnsupportedAction:
		return ErrUnsupportedServiceAction
	case ReasonValidationError:
		return ErrInvalidServiceName
	case ReasonCommandFailed:
		return ErrCommandFailed
	case ReasonCommandTimeout:
		return ErrAgentTimeout
	default:
		return ErrAgentUnavailable
	}
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func ValidateServiceControlRequest(request ServiceControlRequest) error {
	if err := system.ValidateServiceName(request.Service); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidServiceName, err)
	}

	if !isSupportedServiceAction(request.Action) {
		return ErrUnsupportedServiceAction
	}

	return nil
}

func ValidateSocketPath(path string) error {
	if path == "" {
		return ErrInvalidSocketPath
	}
	if !isAbsolutePath(path) {
		return ErrInvalidSocketPath
	}
	return nil
}

func isAbsolutePath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

func isSupportedServiceAction(action string) bool {
	switch action {
	case "start", "stop", "restart":
		return true
	default:
		return false
	}
}
