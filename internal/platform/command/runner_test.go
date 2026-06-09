package command

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunReturnsSuccessfulBoundedOutput(t *testing.T) {
	script := writeScript(t, "printf ok")
	runner := NewTestRunner(time.Second, 16, []string{filepath.Base(script)})

	result, err := runner.Run(context.Background(), script)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if string(result.Stdout) != "ok" {
		t.Fatalf("expected stdout ok, got %q", string(result.Stdout))
	}
}

func TestRunReturnsTimeout(t *testing.T) {
	script := writeScript(t, "sleep 1")
	runner := NewTestRunner(10*time.Millisecond, 16, []string{filepath.Base(script)})

	_, err := runner.Run(context.Background(), script)
	if !errors.Is(err, ErrCommandTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestRunTruncatesOutput(t *testing.T) {
	script := writeScript(t, "printf abcdef")
	runner := NewTestRunner(time.Second, 3, []string{filepath.Base(script)})

	result, err := runner.Run(context.Background(), script)
	if !errors.Is(err, ErrOutputTruncated) {
		t.Fatalf("expected output truncated, got %v", err)
	}
	if string(result.Stdout) != "abc" {
		t.Fatalf("expected truncated stdout abc, got %q", string(result.Stdout))
	}
	if !result.StdoutTruncated {
		t.Fatal("expected stdout truncated marker")
	}
}

func TestRunRejectsDisallowedExecutable(t *testing.T) {
	script := writeScript(t, "printf ok")
	runner := NewTestRunner(time.Second, 16, []string{"other"})

	_, err := runner.Run(context.Background(), script)
	if !errors.Is(err, ErrExecutableNotAllowed) {
		t.Fatalf("expected executable rejection, got %v", err)
	}
}

func writeScript(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "koji-test-command.sh")
	content := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0700); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}
