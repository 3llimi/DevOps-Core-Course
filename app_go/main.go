package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var startTime = time.Now()

type ServiceInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Framework   string `json:"framework"`
}

type SystemInfo struct {
	Hostname        string `json:"hostname"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	Architecture    string `json:"architecture"`
	CPUCount        int    `json:"cpu_count"`
	GoVersion       string `json:"go_version"`
}

type RuntimeInfo struct {
	UptimeSeconds int    `json:"uptime_seconds"`
	UptimeHuman   string `json:"uptime_human"`
	CurrentTime   string `json:"current_time"`
	Timezone      string `json:"timezone"`
}

type RequestInfo struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}

type Endpoint struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

type HomeResponse struct {
	Service   ServiceInfo `json:"service"`
	System    SystemInfo  `json:"system"`
	Runtime   RuntimeInfo `json:"runtime"`
	Request   RequestInfo `json:"request"`
	Endpoints []Endpoint  `json:"endpoints"`
}

type HealthResponse struct {
	Status        string `json:"status"`
	Timestamp     string `json:"timestamp"`
	UptimeSeconds int    `json:"uptime_seconds"`
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getPlatformVersion() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

func getUptime() (int, string) {
	secs := int(time.Since(startTime).Seconds())
	hrs := secs / 3600
	mins := (secs % 3600) / 60
	return secs, fmt.Sprintf("%d hours, %d minutes", hrs, mins)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Printf("404 Not Found: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		http.NotFound(w, r)
		return
	}
	log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	uptime_seconds, uptime_human := getUptime()

	response := HomeResponse{
		Service: ServiceInfo{
			Name:        "devops-info-service",
			Version:     "1.0.0",
			Description: "DevOps course info service",
			Framework:   "net/http",
		},
		System: SystemInfo{
			Hostname:        getHostname(),
			Platform:        runtime.GOOS,
			PlatformVersion: getPlatformVersion(),
			Architecture:    runtime.GOARCH,
			CPUCount:        runtime.NumCPU(),
			GoVersion:       runtime.Version(),
		},
		Runtime: RuntimeInfo{
			UptimeSeconds: uptime_seconds,
			UptimeHuman:   uptime_human,
			CurrentTime:   time.Now().UTC().Format(time.RFC3339),
			Timezone:      "UTC",
		},
		Request: RequestInfo{
			ClientIP:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Method:    r.Method,
			Path:      r.URL.Path,
		},
		Endpoints: []Endpoint{
			{
				Path:        "/",
				Method:      "GET",
				Description: "Service information",
			},
			{
				Path:        "/health",
				Method:      "GET",
				Description: "Health check",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check: %s from %s", r.Method, r.RemoteAddr)
	uptime_seconds, _ := getUptime()
	response := HealthResponse{
		Status:        "healthy",
		Timestamp:     time.Now().UTC().Format(time.RFC3339), // Add .UTC()
		UptimeSeconds: uptime_seconds,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	log.Printf("Starting DevOps Info Service on :%s", port)
	log.Printf("Go version: %s", runtime.Version())
	log.Printf("Platform: %s-%s", runtime.GOOS, runtime.GOARCH)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
