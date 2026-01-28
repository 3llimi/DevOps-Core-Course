# DevOps Info Service (Go)

A Go implementation of the DevOps info service for the bonus task.

## Overview

This service provides the same functionality as the Python version but compiled to a single binary with zero dependencies.

## Prerequisites

- Go 1.21 or higher

## Installation

```bash
cd app_go
go mod download
```

## Running the Application

**Development mode:**
```bash
go run main.go
```

**Build and run binary:**
```bash
go build -o devops-info-service.exe main.go
.\devops-info-service.exe
```

**Custom port:**
```bash
# Windows PowerShell
$env:PORT=3000
go run main.go
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service and system information |
| `/health` | GET | Health check |

## Example Responses

### GET /

```json
{
  "service": {
    "name": "devops-info-service",
    "version": "1.0.0",
    "description": "DevOps course info service",
    "framework": "net/http"
  },
  "system": {
    "hostname": "DESKTOP-ABC123",
    "platform": "windows",
    "platform_version": "Windows-11-10.0.26200-SP0",
    "architecture": "amd64",
    "cpu_count": 8,
    "go_version": "go1.24.0"
  },
  "runtime": {
    "uptime_seconds": 120,
    "uptime_human": "0 hours, 2 minutes",
    "current_time": "2026-01-27T10:30:00Z",
    "timezone": "UTC"
  },
  "request": {
    "client_ip": "127.0.0.1:54321",
    "user_agent": "Mozilla/5.0",
    "method": "GET",
    "path": "/"
  },
  "endpoints": [
    {"path": "/", "method": "GET", "description": "Service information"},
    {"path": "/health", "method": "GET", "description": "Health check"}
  ]
}
```

### GET /health

```json
{
  "status": "healthy",
  "timestamp": "2026-01-27T10:30:00Z",
  "uptime_seconds": 120
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
