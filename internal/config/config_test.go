package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionConfigRequiresServiceAllowlist(t *testing.T) {
	cfg := NewDefaultConfig()

	if err := Validate(cfg); err == nil {
		t.Fatal("expected production service_allowlist requirement")
	}
}

func TestDevConfigUsesNarrowServiceAllowlistDefault(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DevMode = true
	cfg.ServiceAllowlist = DevServiceAllowlist()

	if err := Validate(cfg); err != nil {
		t.Fatalf("expected valid dev config: %v", err)
	}
}

func TestLoadConfigRejectsUnknownField(t *testing.T) {
	path := writeTestConfig(t, "surprise: true\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestLoadConfigRejectsInvalidPort(t *testing.T) {
	path := writeTestConfig(t, "port: 70000\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestLoadConfigAppliesDatabasePath(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "koji.db")
	path := writeTestConfig(t, "database_path: "+databasePath+"\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if cfg.DatabasePath != databasePath {
		t.Fatalf("expected database path %q, got %q", databasePath, cfg.DatabasePath)
	}
}

func TestLoadConfigAppliesAgentSocketPath(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "agent.sock")
	path := writeTestConfig(t, "agent_socket_path: "+socketPath+"\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if cfg.AgentSocketPath != socketPath {
		t.Fatalf("expected agent socket path %q, got %q", socketPath, cfg.AgentSocketPath)
	}
}

func TestLoadConfigRejectsRelativeAgentSocketPath(t *testing.T) {
	path := writeTestConfig(t, "agent_socket_path: relative.sock\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected relative agent socket path error")
	}
}

func TestProductionConfigRejectsRelativeStaticAssetDir(t *testing.T) {
	path := writeTestConfig(t, "static_asset_dir: dist\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected relative static asset dir error")
	}
}

func TestDevConfigAllowsRelativeStaticAssetDir(t *testing.T) {
	path := writeTestConfig(t, "dev_mode: true\nstatic_asset_dir: dist\n")

	_, err := Load(path)
	if err != nil {
		t.Fatalf("expected dev config to allow relative static asset dir: %v", err)
	}
}

func TestLoadConfigAppliesStaticAssetDirAndDevProxyURL(t *testing.T) {
	staticAssetDir := t.TempDir()
	path := writeTestConfig(t, "static_asset_dir: "+staticAssetDir+"\ndev_proxy_url: http://127.0.0.1:5173\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if cfg.StaticAssetDir != staticAssetDir {
		t.Fatalf("expected static asset dir %q, got %q", staticAssetDir, cfg.StaticAssetDir)
	}
	if cfg.DevProxyURL != "http://127.0.0.1:5173" {
		t.Fatalf("unexpected dev proxy url %q", cfg.DevProxyURL)
	}
}

func TestLoadConfigAppliesSessionDurations(t *testing.T) {
	path := writeTestConfig(t, "session_ttl: 8h\nsession_idle_timeout: 20m\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if cfg.SessionTTL.String() != "8h0m0s" {
		t.Fatalf("unexpected session ttl %s", cfg.SessionTTL)
	}
	if cfg.SessionIdleTTL.String() != "20m0s" {
		t.Fatalf("unexpected session idle timeout %s", cfg.SessionIdleTTL)
	}
}

func TestLoadConfigRejectsIdleTimeoutLongerThanSessionTTL(t *testing.T) {
	path := writeTestConfig(t, "session_ttl: 30m\nsession_idle_timeout: 1h\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected idle timeout validation error")
	}
}

func TestLoadConfigAppliesServiceAllowlist(t *testing.T) {
	path := writeBareTestConfig(t, "service_allowlist: ssh.service, kojid.service\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if len(cfg.ServiceAllowlist) != 2 {
		t.Fatalf("expected two allowlisted services, got %#v", cfg.ServiceAllowlist)
	}
}

func TestLoadConfigRejectsInvalidServiceAllowlistEntry(t *testing.T) {
	path := writeBareTestConfig(t, "service_allowlist: ssh/../../bad\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid allowlist config")
	}
}

func TestLoadConfigRejectsEmptyProductionServiceAllowlist(t *testing.T) {
	path := writeBareTestConfig(t, "service_allowlist: []\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected empty production service allowlist rejection")
	}
}

func TestLoadConfigRejectsInvalidProcessVisibilityMode(t *testing.T) {
	path := writeTestConfig(t, "process_visibility_mode: secret\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid process visibility mode rejection")
	}
}

func TestLoadConfigAppliesProcessVisibilityPolicy(t *testing.T) {
	path := writeTestConfig(t, "process_visibility_mode: all\ninclude_command_line: true\nmax_processes: 25\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid process visibility policy: %v", err)
	}
	if cfg.ProcessVisibilityMode != ProcessVisibilityAll {
		t.Fatalf("unexpected process visibility mode %q", cfg.ProcessVisibilityMode)
	}
	if !cfg.IncludeCommandLine {
		t.Fatal("expected command line inclusion")
	}
	if cfg.MaxProcesses != 25 {
		t.Fatalf("expected max processes 25, got %d", cfg.MaxProcesses)
	}
}

func TestLoadConfigRejectsInvalidMaxProcesses(t *testing.T) {
	path := writeTestConfig(t, "max_processes: 0\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid max_processes rejection")
	}
}

func TestLoadAgentDefaultsMutationDisabled(t *testing.T) {
	path := writeBareTestConfig(t, "")

	cfg, err := LoadAgent(path)
	if err != nil {
		t.Fatalf("expected valid disabled agent config: %v", err)
	}
	if cfg.AgentMutationEnabled {
		t.Fatal("expected agent mutation disabled by default")
	}
}

func TestLoadAgentRequiresAllowlistWhenMutationEnabled(t *testing.T) {
	path := writeBareTestConfig(t, "agent_mutation_enabled: true\n")

	_, err := LoadAgent(path)
	if err == nil {
		t.Fatal("expected agent allowlist requirement")
	}
}

func TestLoadAgentRejectsInvalidAllowlistEntry(t *testing.T) {
	path := writeBareTestConfig(t, "agent_mutation_enabled: true\nagent_service_allowlist: bad/name\n")

	_, err := LoadAgent(path)
	if err == nil {
		t.Fatal("expected invalid agent allowlist entry")
	}
}

func TestLoadAgentAppliesCommandBounds(t *testing.T) {
	path := writeBareTestConfig(t, "agent_command_timeout: 5s\nagent_command_output_limit: 4096\n")

	cfg, err := LoadAgent(path)
	if err != nil {
		t.Fatalf("expected valid agent command config: %v", err)
	}
	if cfg.AgentCommandTimeout.String() != "5s" || cfg.AgentCommandOutputLimit != 4096 {
		t.Fatalf("unexpected agent command config: %#v", cfg)
	}
}

func TestLoadAgentParsesYAMLServiceAllowlist(t *testing.T) {
	path := writeBareTestConfig(t, `
agent_mutation_enabled: true
agent_service_allowlist:
  - sshd.service
  - nginx.service
`)

	cfg, err := LoadAgent(path)
	if err != nil {
		t.Fatalf("expected valid agent allowlist config: %v", err)
	}
	if len(cfg.AgentServiceAllowlist) != 2 {
		t.Fatalf("expected two agent services, got %#v", cfg.AgentServiceAllowlist)
	}
}

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()

	if !strings.Contains(content, "service_allowlist:") && !strings.Contains(content, "dev_mode: true") {
		content += "service_allowlist: ssh.service\n"
	}
	return writeBareTestConfig(t, content)
}

func writeBareTestConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "koji.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	return path
}
