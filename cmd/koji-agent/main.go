package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"koji/internal/agent"
	"koji/internal/config"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("Fatal agent error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	flags := parseFlags()
	if flags.ShowVersion {
		fmt.Printf("koji-agent %s\n", version)
		return nil
	}
	cfg, err := loadAgentRuntimeConfig(flags)
	if err != nil {
		return err
	}

	log.Printf("Starting Koji Privileged Agent (koji-agent) %s", version)
	log.Printf("Agent RPC listener binding to Unix socket %s", cfg.AgentSocketPath)
	if !cfg.AgentMutationEnabled {
		log.Println("Service mutation is disabled; service-control returns mutation_disabled.")
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- agent.NewServerWithConfig(cfg).Start(ctx)
	}()

	select {
	case err := <-serverErrCh:
		return err
	case sig := <-shutdownCh:
		log.Printf("Received termination signal %s. koji-agent step down sequence executed.", sig)
		cancel()
	}

	fmt.Println("Clean exit.")
	return nil
}

func loadAgentRuntimeConfig(flags config.Config) (config.Config, error) {
	cfg, err := config.LoadAgent(flags.ConfigPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) || explicitFlag("config") {
			return config.Config{}, err
		}
		cfg = config.NewDefaultConfig()
	}
	applyAgentFlagOverrides(&cfg, flags)
	if err := config.ValidateAgent(cfg); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func applyAgentFlagOverrides(cfg *config.Config, flags config.Config) {
	flag.Visit(func(visited *flag.Flag) {
		switch visited.Name {
		case "socket":
			cfg.AgentSocketPath = flags.AgentSocketPath
		case "mutation-enabled":
			cfg.AgentMutationEnabled = flags.AgentMutationEnabled
		case "service-allowlist":
			cfg.AgentServiceAllowlist = flags.AgentServiceAllowlist
		case "command-timeout":
			cfg.AgentCommandTimeout = flags.AgentCommandTimeout
		case "command-output-limit":
			cfg.AgentCommandOutputLimit = flags.AgentCommandOutputLimit
		}
	})
}

func explicitFlag(name string) bool {
	found := false
	flag.Visit(func(visited *flag.Flag) {
		if visited.Name == name {
			found = true
		}
	})
	return found
}

func parseFlags() config.Config {
	cfg := config.NewDefaultConfig()

	flag.BoolVar(&cfg.ShowVersion, "version", false, "Print version layout and exit")
	flag.StringVar(&cfg.ConfigPath, "config", "/etc/koji/koji.yaml", "Path to agent configuration")
	flag.StringVar(&cfg.AgentSocketPath, "socket", "/run/koji/agent.sock", "Path to koji-agent Unix socket")
	flag.BoolVar(&cfg.AgentMutationEnabled, "mutation-enabled", false, "Enable guarded service mutation")
	flag.Func("service-allowlist", "Comma-separated systemd units eligible for agent mutation", func(value string) error {
		cfg.AgentServiceAllowlist = config.ParseServiceAllowlist(value)
		return nil
	})
	flag.DurationVar(&cfg.AgentCommandTimeout, "command-timeout", cfg.AgentCommandTimeout, "Maximum service mutation command runtime")
	flag.Int64Var(&cfg.AgentCommandOutputLimit, "command-output-limit", cfg.AgentCommandOutputLimit, "Maximum captured service mutation command output")

	flag.Parse()
	return cfg
}
