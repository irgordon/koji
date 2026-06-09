package http

import (
	"errors"
	"net/http"
	"sort"

	"koji/internal/agent"
	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/jobs"
	"koji/internal/system"
)

func (h protectedHandlers) handleServicesList(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.HostServicesRead, "host.services") {
		return
	}

	services := h.serviceAllowlist.services()
	payload := struct {
		Services []system.ServiceStatus `json:"services"`
	}{
		Services: make([]system.ServiceStatus, 0, len(services)),
	}

	for _, unit := range services {
		status, err := system.GetServiceStatus(r.Context(), unit)
		if err != nil {
			continue
		}
		payload.Services = append(payload.Services, status)
	}

	writeJSONValue(w, http.StatusOK, payload)
}

type serviceControlJobResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

func (h protectedHandlers) handleServiceControl(w http.ResponseWriter, r *http.Request) {
	request := serviceControlRequestFromPath(r)
	target := serviceControlTarget(request)

	if !h.requireServiceControlCapability(w, r, target) {
		return
	}

	if err := agent.ValidateServiceControlRequest(request); err != nil {
		if !h.recordServiceControl(w, r, target, audit.OutcomeRejected, serviceControlValidationReason(err)) {
			return
		}
		writeServiceControlValidationError(w, err)
		return
	}

	if !h.serviceAllowlist.allows(request.Service) {
		if !h.recordServiceControl(w, r, target, audit.OutcomeDenied, "service_not_allowlisted") {
			return
		}
		writeJSONError(w, http.StatusForbidden, "Service is not allowlisted")
		return
	}

	job, ok := h.createServiceControlJob(w, r, request, target)
	if !ok {
		return
	}
	writeJSONValue(w, http.StatusAccepted, serviceControlJobResponse{JobID: job.ID, Status: job.Status})
}

type serviceAllowlist map[string]struct{}

func newServiceAllowlist(services []string) serviceAllowlist {
	allowlist := make(serviceAllowlist, len(services))
	for _, service := range services {
		allowlist[service] = struct{}{}
	}
	return allowlist
}

func (allowlist serviceAllowlist) allows(service string) bool {
	_, ok := allowlist[service]
	return ok
}

func (allowlist serviceAllowlist) services() []string {
	services := make([]string, 0, len(allowlist))
	for service := range allowlist {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func serviceControlRequestFromPath(r *http.Request) agent.ServiceControlRequest {
	return agent.ServiceControlRequest{
		Service: r.PathValue("name"),
		Action:  r.PathValue("action"),
	}
}

func serviceControlTarget(request agent.ServiceControlRequest) string {
	return request.Service + ":" + request.Action
}

func serviceControlValidationReason(err error) string {
	switch {
	case errors.Is(err, agent.ErrInvalidServiceName):
		return "invalid_service_name"
	case errors.Is(err, agent.ErrUnsupportedServiceAction):
		return "unsupported_service_action"
	default:
		return "invalid_request"
	}
}

func writeServiceControlValidationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agent.ErrInvalidServiceName):
		writeJSONError(w, http.StatusBadRequest, "Invalid service name")
	case errors.Is(err, agent.ErrUnsupportedServiceAction):
		writeJSONError(w, http.StatusBadRequest, "Unsupported service action")
	default:
		writeJSONError(w, http.StatusBadRequest, "Invalid service control request")
	}
}

func (h protectedHandlers) createServiceControlJob(w http.ResponseWriter, r *http.Request, request agent.ServiceControlRequest, target string) (jobs.Job, bool) {
	principal, ok := principalFromRequest(r)
	if !ok && !h.devMode {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return jobs.Job{}, false
	}
	var userID *int64
	if ok {
		userID = &principal.UserID
	}

	job, err := h.jobs.Create(r.Context(), jobs.CreateRequest{
		CreatedBy: userID,
		Action:    "service." + request.Action,
		Target:    request.Service,
		RequestID: requestID(r),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create job")
		return jobs.Job{}, false
	}

	if !h.recordJobCreated(w, r, userID, job, target) {
		return jobs.Job{}, false
	}
	return job, h.recordJobStatusChanged(w, r, userID, job, target)
}
