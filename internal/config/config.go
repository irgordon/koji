package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"koji/internal/system"
)

var devServiceAllowlist = []string{
	"ssh.service",
	"kojid.service",
}

const (
	ProcessVisibilitySummary = "summary"
	ProcessVisibilityOwner   = "owner"
	ProcessVisibilityAll     = "all"

	DefaultMaxProcesses                  = 200
	DefaultAgentCommandTimeout           = 3 * time.Second
	DefaultAgentCommandOutputLimit       = 64 * 1024
	maxProcessesLimit                    = 1000
	maxAgentCommandOutputLimit     int64 = 1024 * 1024
)

// Config holds all Koji application configuration properties.
// Explicit design matrix with no global hidden states.
type Config struct {
	Port                    int
	DevMode                 bool
	ShowVersion             bool
	ConfigPath              string
	DatabasePath            string
	AgentSocketPath         string
	StaticAssetDir          string
	DevProxyURL             string
	SessionTTL              time.Duration
	SessionIdleTTL          time.Duration
	ServiceAllowlist        []string
	AgentMutationEnabled    bool
	AgentServiceAllowlist   []string
	AgentCommandTimeout     time.Duration
	AgentCommandOutputLimit int64
	ProcessVisibilityMode   string
	IncludeCommandLine      bool
	MaxProcesses            int
}

// NewDefaultConfig returns standard hardcoded system invariants for configuration.
func NewDefaultConfig() Config {
	return Config{
		Port:                    8443,
		DevMode:                 false,
		ShowVersion:             false,
		ConfigPath:              "/etc/koji/koji.yaml",
		DatabasePath:            "/var/lib/koji/koji.db",
		AgentSocketPath:         "/run/koji/agent.sock",
		StaticAssetDir:          "/usr/share/koji/dist",
		DevProxyURL:             "http://localhost:5173",
		SessionTTL:              12 * time.Hour,
		SessionIdleTTL:          30 * time.Minute,
		AgentMutationEnabled:    false,
		AgentCommandTimeout:     DefaultAgentCommandTimeout,
		AgentCommandOutputLimit: DefaultAgentCommandOutputLimit,
		ProcessVisibilityMode:   ProcessVisibilitySummary,
		IncludeCommandLine:      false,
		MaxProcesses:            DefaultMaxProcesses,
	}
}

func DevServiceAllowlist() []string {
	return append([]string(nil), devServiceAllowlist...)
}

func ParseServiceAllowlist(value string) []string {
	return parseServiceAllowlist(value)
}

func Load(path string) (Config, error) {
	cfg := NewDefaultConfig()
	cfg.ConfigPath = path

	content, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer content.Close()

	if err := parseConfig(content, &cfg); err != nil {
		return Config{}, err
	}

	ApplyDefaults(&cfg)

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func LoadAgent(path string) (Config, error) {
	cfg := NewDefaultConfig()
	cfg.ConfigPath = path

	content, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open agent config: %w", err)
	}
	defer content.Close()

	if err := parseConfig(content, &cfg); err != nil {
		return Config{}, err
	}
	if err := ValidateAgent(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Validate(cfg Config) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}
	if cfg.ConfigPath == "" {
		return fmt.Errorf("config path is required")
	}
	if cfg.DatabasePath == "" {
		return fmt.Errorf("database path is required")
	}
	if !filepath.IsAbs(cfg.DatabasePath) {
		return fmt.Errorf("database path must be absolute")
	}
	if cfg.AgentSocketPath == "" {
		return fmt.Errorf("agent socket path is required")
	}
	if !filepath.IsAbs(cfg.AgentSocketPath) {
		return fmt.Errorf("agent socket path must be absolute")
	}
	if cfg.StaticAssetDir == "" {
		return fmt.Errorf("static asset dir is required")
	}
	if !cfg.DevMode && !filepath.IsAbs(cfg.StaticAssetDir) {
		return fmt.Errorf("static asset dir must be absolute in production")
	}
	if cfg.DevProxyURL != "" {
		if err := validateURL(cfg.DevProxyURL); err != nil {
			return err
		}
	}
	if cfg.SessionTTL <= 0 {
		return fmt.Errorf("session ttl must be positive")
	}
	if cfg.SessionIdleTTL <= 0 {
		return fmt.Errorf("session idle timeout must be positive")
	}
	if cfg.SessionIdleTTL > cfg.SessionTTL {
		return fmt.Errorf("session idle timeout must not exceed session ttl")
	}
	if err := validateServiceAllowlist(cfg); err != nil {
		return err
	}
	if err := validateProcessVisibility(cfg); err != nil {
		return err
	}
	return nil
}

func ValidateAgent(cfg Config) error {
	if cfg.ConfigPath == "" {
		return fmt.Errorf("config path is required")
	}
	if cfg.AgentSocketPath == "" {
		return fmt.Errorf("agent socket path is required")
	}
	if !filepath.IsAbs(cfg.AgentSocketPath) {
		return fmt.Errorf("agent socket path must be absolute")
	}
	if cfg.AgentCommandTimeout <= 0 {
		return fmt.Errorf("agent command timeout must be positive")
	}
	if cfg.AgentCommandOutputLimit <= 0 || cfg.AgentCommandOutputLimit > maxAgentCommandOutputLimit {
		return fmt.Errorf("agent command output limit must be between 1 and %d", maxAgentCommandOutputLimit)
	}
	return validateAgentServiceAllowlist(cfg)
}

func ApplyDefaults(cfg *Config) {
	if cfg.DevMode && len(cfg.ServiceAllowlist) == 0 {
		cfg.ServiceAllowlist = DevServiceAllowlist()
	}
}

func validateServiceAllowlist(cfg Config) error {
	if !cfg.DevMode && len(cfg.ServiceAllowlist) == 0 {
		return fmt.Errorf("service_allowlist is required in production")
	}
	for _, service := range cfg.ServiceAllowlist {
		if err := system.ValidateServiceName(service); err != nil {
			return fmt.Errorf("invalid service_allowlist entry %q", service)
		}
	}
	return nil
}

func validateAgentServiceAllowlist(cfg Config) error {
	if cfg.AgentMutationEnabled && len(cfg.AgentServiceAllowlist) == 0 {
		return fmt.Errorf("agent_service_allowlist is required when agent mutation is enabled")
	}
	for _, service := range cfg.AgentServiceAllowlist {
		if err := system.ValidateServiceName(service); err != nil {
			return fmt.Errorf("invalid agent_service_allowlist entry %q", service)
		}
	}
	return nil
}

func validateProcessVisibility(cfg Config) error {
	switch cfg.ProcessVisibilityMode {
	case ProcessVisibilitySummary, ProcessVisibilityOwner, ProcessVisibilityAll:
	default:
		return fmt.Errorf("invalid process_visibility_mode %q", cfg.ProcessVisibilityMode)
	}
	if cfg.MaxProcesses < 1 || cfg.MaxProcesses > maxProcessesLimit {
		return fmt.Errorf("max_processes must be between 1 and %d", maxProcessesLimit)
	}
	return nil
}

func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid dev proxy url")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("dev proxy url must be absolute")
	}
	return nil
}

func parseConfig(content *os.File, cfg *Config) error {
	scanner := bufio.NewScanner(content)
	seen := map[string]bool{}
	pendingListKey := ""

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if isIgnoredConfigLine(line) {
			continue
		}
		if isListItem(line) {
			if pendingListKey == "" {
				return fmt.Errorf("parse config line %d: unexpected list item", lineNumber)
			}
			if err := appendConfigListItem(cfg, pendingListKey, line); err != nil {
				return fmt.Errorf("parse config line %d: %w", lineNumber, err)
			}
			continue
		}
		pendingListKey = ""

		key, value, err := splitConfigLine(line)
		if err != nil {
			return fmt.Errorf("parse config line %d: %w", lineNumber, err)
		}
		if seen[key] {
			return fmt.Errorf("duplicate config field %q", key)
		}
		seen[key] = true

		if err := applyConfigField(cfg, key, value); err != nil {
			return fmt.Errorf("parse config line %d: %w", lineNumber, err)
		}
		if isListConfigField(key) && value == "" {
			pendingListKey = key
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan config: %w", err)
	}
	return nil
}

func isListItem(line string) bool {
	return strings.HasPrefix(line, "- ")
}

func isListConfigField(key string) bool {
	return key == "service_allowlist" || key == "agent_service_allowlist"
}

func appendConfigListItem(cfg *Config, key string, line string) error {
	value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "- ")), `"'`)
	if value == "" {
		return fmt.Errorf("empty list item")
	}
	switch key {
	case "service_allowlist":
		cfg.ServiceAllowlist = append(cfg.ServiceAllowlist, value)
	case "agent_service_allowlist":
		cfg.AgentServiceAllowlist = append(cfg.AgentServiceAllowlist, value)
	default:
		return fmt.Errorf("unsupported list field %q", key)
	}
	return nil
}

func isIgnoredConfigLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func splitConfigLine(line string) (string, string, error) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", fmt.Errorf("expected key: value")
	}

	key = strings.TrimSpace(key)
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	if key == "" {
		return "", "", fmt.Errorf("config field name is required")
	}
	return key, value, nil
}

func applyConfigField(cfg *Config, key string, value string) error {
	switch key {
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port value")
		}
		cfg.Port = port
	case "dev_mode":
		devMode, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid dev_mode value")
		}
		cfg.DevMode = devMode
	case "database_path":
		cfg.DatabasePath = value
	case "agent_socket_path":
		cfg.AgentSocketPath = value
	case "agent_mutation_enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid agent_mutation_enabled value")
		}
		cfg.AgentMutationEnabled = enabled
	case "agent_service_allowlist":
		cfg.AgentServiceAllowlist = parseServiceAllowlist(value)
	case "agent_command_timeout":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid agent_command_timeout value")
		}
		cfg.AgentCommandTimeout = duration
	case "agent_command_output_limit":
		limit, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid agent_command_output_limit value")
		}
		cfg.AgentCommandOutputLimit = limit
	case "static_asset_dir":
		cfg.StaticAssetDir = value
	case "dev_proxy_url":
		cfg.DevProxyURL = value
	case "session_ttl":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid session_ttl value")
		}
		cfg.SessionTTL = duration
	case "session_idle_timeout":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid session_idle_timeout value")
		}
		cfg.SessionIdleTTL = duration
	case "service_allowlist":
		cfg.ServiceAllowlist = parseServiceAllowlist(value)
	case "process_visibility_mode":
		cfg.ProcessVisibilityMode = value
	case "include_command_line":
		includeCommandLine, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid include_command_line value")
		}
		cfg.IncludeCommandLine = includeCommandLine
	case "max_processes":
		maxProcesses, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_processes value")
		}
		cfg.MaxProcesses = maxProcesses
	default:
		return fmt.Errorf("unknown config field %q", key)
	}
	return nil
}

func parseServiceAllowlist(value string) []string {
	value = strings.Trim(value, "[]")
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	services := make([]string, 0, len(parts))
	for _, part := range parts {
		service := strings.Trim(strings.TrimSpace(part), `"'`)
		if service != "" {
			services = append(services, service)
		}
	}
	return services
}
