package http

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"koji/internal/agent"
	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/jobs"
)

type recordingServiceController struct {
	requests []agent.ServiceControlRequest
	err      error
}

func (c *recordingServiceController) ControlService(ctx context.Context, request agent.ServiceControlRequest) error {
	c.requests = append(c.requests, request)
	return c.err
}

func TestServiceControlRejectsInvalidServiceNameBeforeAgent(t *testing.T) {
	controller := &recordingServiceController{}
	response := exerciseServiceControl(t, controller, "ssh/../../bad", "restart")

	if response.Code != 400 {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
	if len(controller.requests) != 0 {
		t.Fatalf("expected no agent call, got %d", len(controller.requests))
	}
}

func TestServiceControlRejectsUnsupportedActionBeforeAgent(t *testing.T) {
	controller := &recordingServiceController{}
	response := exerciseServiceControl(t, controller, "ssh.service", "reload")

	if response.Code != 400 {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
	if len(controller.requests) != 0 {
		t.Fatalf("expected no agent call, got %d", len(controller.requests))
	}
}

func TestServiceControlCreatesQueuedJob(t *testing.T) {
	controller := &recordingServiceController{}
	response := exerciseServiceControl(t, controller, "ssh.service", "restart")

	if response.Code != 202 {
		t.Fatalf("expected status 202, got %d", response.Code)
	}
	if len(controller.requests) != 0 {
		t.Fatalf("expected no synchronous agent call, got %d", len(controller.requests))
	}
	if !strings.Contains(response.Body.String(), `"status":"queued"`) {
		t.Fatalf("expected queued job response, got %q", response.Body.String())
	}
}

func TestServiceControlDeniesNonAllowlistedService(t *testing.T) {
	controller := &recordingServiceController{}
	response := exerciseServiceControl(t, controller, "nginx.service", "restart")

	if response.Code != 403 {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
	if len(controller.requests) != 0 {
		t.Fatalf("expected no agent call, got %d", len(controller.requests))
	}
}

func exerciseServiceControl(t *testing.T, controller agent.ServiceController, service string, action string) *httptest.ResponseRecorder {
	t.Helper()

	fixture := newTestFixture(t)
	defer fixture.cleanup()

	request := httptest.NewRequest("POST", "/api/services/"+service+"/"+action, nil)
	request.SetPathValue("name", service)
	request.SetPathValue("action", action)

	response := httptest.NewRecorder()
	handler := protectedHandlers{
		caps:             caps.NewStore(fixture.database),
		audit:            audit.NewStore(fixture.database),
		devMode:          true,
		serviceAllowlist: newServiceAllowlist([]string{"ssh.service"}),
		jobs:             jobs.NewStore(fixture.database),
	}
	handler.handleServiceControl(response, request)
	return response
}
