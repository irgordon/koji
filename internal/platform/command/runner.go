package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	DefaultTimeout     = 3 * time.Second
	DefaultOutputLimit = 64 * 1024
	ServiceManager     = "systemctl"
)

var (
	ErrExecutableNotAllowed = errors.New("command executable not allowed")
	ErrCommandTimeout       = errors.New("command timed out")
	ErrOutputTruncated      = errors.New("command output truncated")
	ErrCommandFailed        = errors.New("command failed")
)

type Runner struct {
	timeout     time.Duration
	outputLimit int64
	allowed     map[string]bool
}

type Result struct {
	Stdout          []byte
	Stderr          []byte
	StdoutTruncated bool
	StderrTruncated bool
}

type CommandRunner interface {
	Run(ctx context.Context, executable string, args ...string) (Result, error)
}

func RunServiceStatus(ctx context.Context, runner Runner, name string) (Result, error) {
	return runner.Run(ctx, ServiceManager, "show", name, "--property=ActiveState,SubState")
}

func RunServiceMutation(ctx context.Context, runner CommandRunner, action string, name string) (Result, error) {
	return runner.Run(ctx, ServiceManager, action, name)
}

func NewReadOnlyRunner() Runner {
	return Runner{
		timeout:     DefaultTimeout,
		outputLimit: DefaultOutputLimit,
		allowed: map[string]bool{
			ServiceManager: true,
		},
	}
}

func NewAgentMutationRunner(timeout time.Duration, outputLimit int64) Runner {
	return Runner{
		timeout:     effectiveTimeout(timeout),
		outputLimit: effectiveOutputLimit(outputLimit),
		allowed: map[string]bool{
			ServiceManager: true,
		},
	}
}

func NewTestRunner(timeout time.Duration, outputLimit int64, allowed []string) Runner {
	return Runner{
		timeout:     timeout,
		outputLimit: outputLimit,
		allowed:     allowedSet(allowed),
	}
}

func effectiveTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultTimeout
	}
	return timeout
}

func effectiveOutputLimit(outputLimit int64) int64 {
	if outputLimit <= 0 {
		return DefaultOutputLimit
	}
	return outputLimit
}

func (r Runner) Run(ctx context.Context, executable string, args ...string) (Result, error) {
	if !r.executableAllowed(executable) {
		return Result{}, ErrExecutableNotAllowed
	}

	runCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var stdout, stderr limitedBuffer
	stdout.limit = r.outputLimit
	stderr.limit = r.outputLimit

	cmd := exec.CommandContext(runCtx, executable, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{
		Stdout:          stdout.Bytes(),
		Stderr:          stderr.Bytes(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return result, ErrCommandTimeout
	}
	if err != nil {
		return result, ErrCommandFailed
	}
	if result.StdoutTruncated || result.StderrTruncated {
		return result, ErrOutputTruncated
	}
	return result, nil
}

func (r Runner) executableAllowed(executable string) bool {
	name := filepath.Base(executable)
	return r.allowed[name]
}

func allowedSet(names []string) map[string]bool {
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	return allowed
}

type limitedBuffer struct {
	buffer    bytes.Buffer
	limit     int64
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		b.truncated = true
		return len(p), nil
	}

	remaining := b.limit - int64(b.buffer.Len())
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}

	if int64(len(p)) > remaining {
		_, _ = b.buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}

	_, _ = b.buffer.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) Bytes() []byte {
	return b.buffer.Bytes()
}

func (b *limitedBuffer) Truncated() bool {
	return b.truncated
}

func SafeError(err error) error {
	switch {
	case errors.Is(err, ErrExecutableNotAllowed):
		return ErrExecutableNotAllowed
	case errors.Is(err, ErrCommandTimeout):
		return ErrCommandTimeout
	case errors.Is(err, ErrOutputTruncated):
		return ErrOutputTruncated
	case errors.Is(err, ErrCommandFailed):
		return ErrCommandFailed
	default:
		return fmt.Errorf("%w", ErrCommandFailed)
	}
}

var _ io.Writer = (*limitedBuffer)(nil)
