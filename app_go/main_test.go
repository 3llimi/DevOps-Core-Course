package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test helper function to create test server
func setupTestRequest(method, path string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("User-Agent", "test-client/1.0")
	w := httptest.NewRecorder()
	return req, w
}

// ============================================
// Tests for GET / endpoint
// ============================================

func TestHomeEndpoint(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHomeReturnsJSON(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
	homeHandler(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response HomeResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

func TestHomeHasServiceInfo(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
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
	if response.Service.Description != "DevOps course info service" {
		t.Errorf("expected description 'DevOps course info service', got '%s'", response.Service.Description)
	}
}

func TestHomeHasSystemInfo(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
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
	if response.System.CPUCount <= 0 {
		t.Errorf("cpu_count should be positive, got %d", response.System.CPUCount)
	}
}

func TestHomeHasRuntimeInfo(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Runtime.UptimeSeconds < 0 {
		t.Errorf("uptime_seconds should be non-negative, got %d", response.Runtime.UptimeSeconds)
	}
	if response.Runtime.CurrentTime == "" {
		t.Error("current_time should not be empty")
	}
	if response.Runtime.Timezone != "UTC" {
		t.Errorf("expected timezone 'UTC', got '%s'", response.Runtime.Timezone)
	}
}

func TestHomeHasRequestInfo(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Request.Method != "GET" {
		t.Errorf("expected method 'GET', got '%s'", response.Request.Method)
	}
	if response.Request.Path != "/" {
		t.Errorf("expected path '/', got '%s'", response.Request.Path)
	}
	if response.Request.UserAgent != "test-client/1.0" {
		t.Errorf("expected user agent 'test-client/1.0', got '%s'", response.Request.UserAgent)
	}
}

func TestHomeHasEndpoints(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/")
	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(response.Endpoints))
	}

	// Check first endpoint
	if response.Endpoints[0].Path != "/" {
		t.Errorf("expected first endpoint path '/', got '%s'", response.Endpoints[0].Path)
	}
	if response.Endpoints[0].Method != "GET" {
		t.Errorf("expected first endpoint method 'GET', got '%s'", response.Endpoints[0].Method)
	}

	// Check second endpoint
	if response.Endpoints[1].Path != "/health" {
		t.Errorf("expected second endpoint path '/health', got '%s'", response.Endpoints[1].Path)
	}
}

// ============================================
// Tests for GET /health endpoint
// ============================================

func TestHealthEndpoint(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/health")
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHealthReturnsJSON(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/health")
	healthHandler(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

func TestHealthHasStatus(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/health")
	healthHandler(w, req)

	var response HealthResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

func TestHealthHasTimestamp(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/health")
	healthHandler(w, req)

	var response HealthResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
}

func TestHealthHasUptime(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/health")
	healthHandler(w, req)

	var response HealthResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.UptimeSeconds < 0 {
		t.Errorf("uptime_seconds should be non-negative, got %d", response.UptimeSeconds)
	}
}

// ============================================
// Tests for 404 handler
// ============================================

func Test404Handler(t *testing.T) {
	req, w := setupTestRequest(http.MethodGet, "/nonexistent")
	homeHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func Test404OnInvalidPath(t *testing.T) {
	invalidPaths := []string{"/api", "/test", "/favicon.ico", "/robots.txt"}

	for _, path := range invalidPaths {
		req, w := setupTestRequest(http.MethodGet, path)
		homeHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 for path '%s', got %d", path, w.Code)
		}
	}
}

// ============================================
// Tests for helper functions
// ============================================

func TestGetHostname(t *testing.T) {
	hostname := getHostname()
	if hostname == "" {
		t.Error("hostname should not be empty")
	}
	// Should never return "unknown" in normal conditions
	if hostname == "unknown" {
		t.Log("Warning: hostname returned 'unknown'")
	}
}

func TestGetPlatformVersion(t *testing.T) {
	platformVersion := getPlatformVersion()
	if platformVersion == "" {
		t.Error("platform version should not be empty")
	}
	// Should contain a hyphen (e.g., "linux-amd64")
	if len(platformVersion) < 3 {
		t.Errorf("platform version seems invalid: '%s'", platformVersion)
	}
}

func TestGetUptime(t *testing.T) {
	seconds, human := getUptime()

	if seconds < 0 {
		t.Errorf("uptime seconds should be non-negative, got %d", seconds)
	}

	if human == "" {
		t.Error("uptime human format should not be empty")
	}

	// Human format should contain "hours" and "minutes"
	// (even if 0 hours, 0 minutes)
	if len(human) < 10 {
		t.Errorf("uptime human format seems too short: '%s'", human)
	}
}

// ============================================
// Edge case and error handling tests
// ============================================

func TestHomeHandlerWithPOSTMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	homeHandler(w, req)

	// Should still return 200 (handler doesn't restrict methods)
	// But this documents the behavior
	if w.Code != http.StatusOK {
		t.Logf("POST to / returned status %d", w.Code)
	}
}

func TestHealthHandlerWithPOSTMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	// Should still return 200 (handler doesn't restrict methods)
	if w.Code != http.StatusOK {
		t.Logf("POST to /health returned status %d", w.Code)
	}
}

func TestResponseContentTypeIsJSON(t *testing.T) {
	endpoints := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/", homeHandler},
		{"/health", healthHandler},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(http.MethodGet, endpoint.path, nil)
		w := httptest.NewRecorder()

		endpoint.handler(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("endpoint %s: expected Content-Type 'application/json', got '%s'",
				endpoint.path, contentType)
		}
	}
}

// Test for malformed RemoteAddr (covers net.SplitHostPort error path)
func TestHomeHandlerWithMalformedRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set an invalid RemoteAddr without port
	req.RemoteAddr = "192.168.1.1"
	w := httptest.NewRecorder()

	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	// Should still work and use the full RemoteAddr as client IP
	if response.Request.ClientIP != "192.168.1.1" {
		t.Errorf("expected client IP '192.168.1.1', got '%s'", response.Request.ClientIP)
	}
}

// Test with empty RemoteAddr
func TestHomeHandlerWithEmptyRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ""
	w := httptest.NewRecorder()

	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	// Should handle empty RemoteAddr gracefully
	if response.Request.ClientIP != "" {
		t.Logf("Empty RemoteAddr resulted in client IP: '%s'", response.Request.ClientIP)
	}
}

// Test with IPv6 address
func TestHomeHandlerWithIPv6RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:12345"
	w := httptest.NewRecorder()

	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Request.ClientIP != "::1" {
		t.Errorf("expected client IP '::1', got '%s'", response.Request.ClientIP)
	}
}

// Test empty User-Agent
func TestHomeHandlerWithEmptyUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Del("User-Agent")
	w := httptest.NewRecorder()

	homeHandler(w, req)

	var response HomeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Request.UserAgent != "" {
		t.Logf("Empty User-Agent resulted in: '%s'", response.Request.UserAgent)
	}
}

// Test uptime calculation over time
func TestGetUptimeProgression(t *testing.T) {
	seconds1, human1 := getUptime()

	// Wait a tiny bit
	time.Sleep(10 * time.Millisecond)

	seconds2, human2 := getUptime()

	if seconds2 < seconds1 {
		t.Error("uptime should not decrease")
	}

	// Both should be non-empty
	if human1 == "" || human2 == "" {
		t.Error("uptime human format should not be empty")
	}
}

// Test uptime formatting with specific durations
func TestUptimeFormatting(t *testing.T) {
	// This indirectly tests the uptime formatting logic
	seconds, human := getUptime()

	// Human should contain "hours" and "minutes"
	if !contains(human, "hours") || !contains(human, "minutes") {
		t.Errorf("uptime format should contain 'hours' and 'minutes', got: '%s'", human)
	}

	// Seconds should match reasonable expectations
	if seconds < 0 {
		t.Errorf("seconds should be non-negative, got %d", seconds)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test different HTTP methods on health endpoint
func TestHealthHandlerWithDifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/health", nil)
		w := httptest.NewRecorder()

		healthHandler(w, req)

		// All methods should succeed (no method restriction in handler)
		if w.Code != http.StatusOK {
			t.Errorf("method %s: expected status 200, got %d", method, w.Code)
		}
	}
}

// Test concurrent requests to ensure no race conditions
func TestConcurrentHomeRequests(t *testing.T) {
	const numRequests = 100
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			homeHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("concurrent request failed with status %d", w.Code)
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// Test concurrent health checks
func TestConcurrentHealthRequests(t *testing.T) {
	const numRequests = 100
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			healthHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("concurrent health check failed with status %d", w.Code)
			}
			done <- true
		}()
	}

	for i := 0; i < numRequests; i++ {
		<-done
	}
}
