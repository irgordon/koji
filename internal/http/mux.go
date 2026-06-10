package http

import (
	"database/sql"
	"net/http"
	"time"

	"koji/internal/agent"
	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
	"koji/internal/config"
	"koji/internal/identity"
	"koji/internal/jobs"
	"koji/internal/observability"
	"koji/internal/system"
)

type protectedHandlers struct {
	caps             *caps.Store
	audit            *audit.Store
	devMode          bool
	serviceAllowlist serviceAllowlist
	processPolicy    processVisibilityPolicy
	jobs             *jobs.Store
	identity         *identity.Store
	magicTokenTTL    time.Duration
	metrics          *observability.Registry
	database         *sql.DB
}

type routeDependencies struct {
	config      config.Config
	authStore   *auth.Store
	auditStore  *audit.Store
	protected   protectedHandlers
	operational operationalHandlers
	probe       *system.Probe
	agentClient agent.ServiceController
}

func NewMuxRouter(cfg config.Config, database *sql.DB) (http.Handler, error) {
	return NewMuxRouterWithAgent(cfg, database, agent.NewClient(cfg.AgentSocketPath))
}

func NewMuxRouterWithAgent(cfg config.Config, database *sql.DB, agentClient agent.ServiceController) (http.Handler, error) {
	deps, err := newRouteDependencies(cfg, database, agentClient, observability.DefaultRegistry())
	if err != nil {
		return nil, err
	}

	mux, err := registerRoutes(deps)
	if err != nil {
		return nil, err
	}

	return applyMiddlewareChain(mux, deps.authStore, cfg.DevMode), nil
}

func newRouteDependencies(
	cfg config.Config,
	database *sql.DB,
	agentClient agent.ServiceController,
	metrics *observability.Registry,
) (routeDependencies, error) {
	authStore := auth.NewStoreWithPolicy(database, auth.SessionPolicy{
		TTL:         cfg.SessionTTL,
		IdleTimeout: cfg.SessionIdleTTL,
	})
	auditStore := audit.NewStoreWithMetrics(database, metrics)
	protected := protectedHandlers{
		caps:             caps.NewStore(database),
		audit:            auditStore,
		devMode:          cfg.DevMode,
		serviceAllowlist: newServiceAllowlist(cfg.ServiceAllowlist),
		processPolicy:    newProcessVisibilityPolicy(cfg),
		jobs:             jobs.NewStoreWithMetrics(database, metrics),
		identity:         identity.NewStore(database),
		magicTokenTTL:    cfg.MagicTokenTTL,
		metrics:          metrics,
		database:         database,
	}

	probe, err := system.NewProbe()
	if err != nil {
		return routeDependencies{}, err
	}

	return routeDependencies{
		config:     cfg,
		authStore:  authStore,
		auditStore: auditStore,
		protected:  protected,
		operational: operationalHandlers{
			database:        database,
			agentSocketPath: cfg.AgentSocketPath,
			metrics:         metrics,
		},
		probe:       probe,
		agentClient: agentClient,
	}, nil
}
