package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"koji/internal/agent"
	"koji/internal/audit"
	"koji/internal/config"
	"koji/internal/db"
	kojihttp "koji/internal/http"
	"koji/internal/jobs"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("Fatal application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	flags := parseFlags()

	if flags.ShowVersion {
		fmt.Printf("kojid %s\n", version)
		return nil
	}

	log.Printf("Starting Koji API Daemon (kojid) %s", version)

	cfg, err := loadRuntimeConfig(flags)
	if err != nil {
		return err
	}

	database, err := initializeDatabase(cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	srv, err := kojihttp.New(cfg, database)
	if err != nil {
		return fmt.Errorf("failed to initialize server platform: %w", err)
	}

	runtimeCtx, stopRuntime := context.WithCancel(context.Background())
	defer stopRuntime()

	workerErrCh := startJobWorker(runtimeCtx, cfg, database)
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP listener binding to address %s", srv.Addr())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrCh:
		stopRuntime()
		return fmt.Errorf("unrecoverable startup error: %w", err)
	case err := <-workerErrCh:
		stopRuntime()
		return fmt.Errorf("job worker stopped unexpectedly: %w", err)
	case sig := <-shutdownCh:
		log.Printf("Received OS signal %s. Initiating graceful drain sequence...", sig)
	}

	stopRuntime()
	shutdownCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful termination failed to complete safely: %w", err)
	}

	log.Println("All transport connections drained. kojid process exiting cleanly.")
	return nil
}

func startJobWorker(ctx context.Context, cfg config.Config, database *sql.DB) <-chan error {
	errCh := make(chan error, 1)
	worker := jobs.NewWorker(
		jobs.NewStore(database),
		agent.NewClient(cfg.AgentSocketPath),
		audit.NewStore(database),
		jobs.DefaultWorkerPollInterval,
	)
	go func() {
		log.Println("Job worker started.")
		if err := worker.Start(ctx); err != nil {
			errCh <- err
		}
	}()
	return errCh
}

func loadRuntimeConfig(flags config.Config) (config.Config, error) {
	cfg, err := config.Load(flags.ConfigPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("load runtime config: %w", err)
	}
	applyFlagOverrides(&cfg, flags)
	config.ApplyDefaults(&cfg)
	if err := config.Validate(cfg); err != nil {
		return config.Config{}, fmt.Errorf("validate runtime config: %w", err)
	}
	return cfg, nil
}

func applyFlagOverrides(cfg *config.Config, flags config.Config) {
	flag.Visit(func(visited *flag.Flag) {
		switch visited.Name {
		case "port":
			cfg.Port = flags.Port
		case "dev":
			cfg.DevMode = flags.DevMode
		case "database":
			cfg.DatabasePath = flags.DatabasePath
		case "agent-socket":
			cfg.AgentSocketPath = flags.AgentSocketPath
		case "static-dir":
			cfg.StaticAssetDir = flags.StaticAssetDir
		case "dev-proxy":
			cfg.DevProxyURL = flags.DevProxyURL
		case "session-ttl":
			cfg.SessionTTL = flags.SessionTTL
		case "session-idle-timeout":
			cfg.SessionIdleTTL = flags.SessionIdleTTL
		case "service-allowlist":
			cfg.ServiceAllowlist = flags.ServiceAllowlist
		case "process-visibility":
			cfg.ProcessVisibilityMode = flags.ProcessVisibilityMode
		case "include-command-line":
			cfg.IncludeCommandLine = flags.IncludeCommandLine
		case "max-processes":
			cfg.MaxProcesses = flags.MaxProcesses
		}
	})
}

func initializeDatabase(cfg config.Config) (*sql.DB, error) {
	database, err := db.Open(context.Background(), cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return database, nil
}

func parseFlags() config.Config {
	cfg := config.NewDefaultConfig()

	flag.IntVar(&cfg.Port, "port", 8443, "HTTP listen port")
	flag.BoolVar(&cfg.DevMode, "dev", false, "Enable development mode (bypasses static asset security layers)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Print version layout and exit")
	flag.StringVar(&cfg.ConfigPath, "config", "/etc/koji/koji.yaml", "Path to system configuration")
	flag.StringVar(&cfg.DatabasePath, "database", "/var/lib/koji/koji.db", "Path to data store matrix")
	flag.StringVar(&cfg.AgentSocketPath, "agent-socket", "/run/koji/agent.sock", "Path to koji-agent Unix socket")
	flag.StringVar(&cfg.StaticAssetDir, "static-dir", "/usr/share/koji/dist", "Path to production frontend assets")
	flag.StringVar(&cfg.DevProxyURL, "dev-proxy", "http://localhost:5173", "Development frontend proxy URL")
	flag.DurationVar(&cfg.SessionTTL, "session-ttl", cfg.SessionTTL, "Maximum session lifetime")
	flag.DurationVar(&cfg.SessionIdleTTL, "session-idle-timeout", cfg.SessionIdleTTL, "Maximum idle session lifetime")
	flag.Func("service-allowlist", "Comma-separated systemd units eligible for service APIs", func(value string) error {
		cfg.ServiceAllowlist = config.ParseServiceAllowlist(value)
		return nil
	})
	flag.StringVar(&cfg.ProcessVisibilityMode, "process-visibility", cfg.ProcessVisibilityMode, "Process visibility mode: summary, owner, or all")
	flag.BoolVar(&cfg.IncludeCommandLine, "include-command-line", cfg.IncludeCommandLine, "Include full process command lines")
	flag.IntVar(&cfg.MaxProcesses, "max-processes", cfg.MaxProcesses, "Maximum processes returned by the API")

	flag.Parse()
	return cfg
}
