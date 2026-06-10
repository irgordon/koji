package http

import "net/http"

func registerRoutes(deps routeDependencies) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	registerOperationalRoutes(mux, deps)
	registerAPIRoutes(mux, deps)
	if err := registerStaticRoutes(mux, deps.config); err != nil {
		return nil, err
	}

	return mux, nil
}

func registerOperationalRoutes(mux *http.ServeMux, deps routeDependencies) {
	mux.HandleFunc("GET /healthz", deps.operational.handleHealth)
	mux.HandleFunc("GET /readyz", deps.operational.handleReady)
}

func registerAPIRoutes(mux *http.ServeMux, deps routeDependencies) {
	mux.HandleFunc("GET /api/v1/metrics", deps.protected.handleMetricsFetch(deps.probe))
	mux.HandleFunc("GET /api/v1/disk", deps.protected.handleDiskFetch)
	mux.HandleFunc("GET /api/v1/services", deps.protected.handleServicesList)
	mux.HandleFunc("POST /api/services/{name}/{action}", deps.protected.handleServiceControl)
	mux.HandleFunc("GET /api/v1/processes", deps.protected.handleProcessesList)
	mux.HandleFunc("GET /api/activity", deps.protected.handleActivityList)
	mux.HandleFunc("GET /api/observability/metrics", deps.protected.handleObservabilityMetrics)
	mux.HandleFunc("GET /api/jobs", deps.protected.handleJobsList)
	mux.HandleFunc("GET /api/jobs/{id}", deps.protected.handleJobDetail)
	mux.HandleFunc("POST /api/jobs/{id}/approve", deps.protected.handleJobApprove)
	mux.HandleFunc("POST /api/jobs/{id}/reject", deps.protected.handleJobReject)
	mux.HandleFunc("GET /api/admin/users", deps.protected.handleAdminUsersList)
	mux.HandleFunc("POST /api/admin/users", deps.protected.handleAdminUserCreate)
	mux.HandleFunc("POST /api/admin/users/{id}/disable", deps.protected.handleAdminUserDisable)
	mux.HandleFunc("POST /api/admin/users/{id}/enable", deps.protected.handleAdminUserEnable)
	mux.HandleFunc("GET /api/admin/users/{id}/capabilities", deps.protected.handleAdminUserCapabilities)
	mux.HandleFunc("POST /api/admin/users/{id}/capabilities", deps.protected.handleAdminCapabilityGrant)
	mux.HandleFunc("DELETE /api/admin/users/{id}/capabilities/{capability}", deps.protected.handleAdminCapabilityRevoke)
	mux.HandleFunc("POST /api/admin/users/{id}/magic-token", deps.protected.handleAdminMagicTokenIssue)
	mux.HandleFunc("POST /api/bootstrap", handleBootstrap(deps.authStore, deps.auditStore, deps.config.DevMode))
	mux.HandleFunc("POST /api/login", handleLogin(deps.authStore, deps.auditStore, deps.config.DevMode))
	mux.HandleFunc("POST /api/login/magic-token", handleMagicTokenLogin(deps.authStore, deps.auditStore, deps.config.DevMode))
	mux.HandleFunc("POST /api/logout", handleLogout(deps.authStore, deps.auditStore))
	mux.HandleFunc("GET /api/session", handleSessionStatus(deps.authStore))
	mux.HandleFunc("/api/", handleAPINotImplemented)
}

func handleAPINotImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "API endpoint not implemented", http.StatusNotImplemented)
}
