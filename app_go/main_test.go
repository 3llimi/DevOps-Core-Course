package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHomeReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	var response HomeResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

func TestHomeHasServiceInfo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Service.Name != "devops-info-service" {
		t.Errorf("expected service name 'devops-info-service', got '%s'", response.Service.Name)
	}
	if response.Service.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", response.Service.Version)
	}
	if response.Service.Framework != "Go net/http" {
		t.Errorf("expected framework 'Go net/http', got '%s'", response.Service.Framework)
	}
}

func TestHomeHasSystemInfo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.System.Hostname == "" {
		t.Error("hostname should not be empty")
	}
	if response.System.Platform == "" {
		t.Error("platform should not be empty")
	}
	if response.System.GoVersion == "" {
		t.Error("go_version should not be empty")
	}
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

func TestHealthHasStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	var response HealthResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

func TestHealthHasUptime(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	var response HealthResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.UptimeSeconds < 0 {
		t.Errorf("uptime_seconds should be non-negative, got %d", response.UptimeSeconds)
	}
}

func Test404Handler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
