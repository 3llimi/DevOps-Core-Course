# Lab 1 Bonus — Go Implementation

## Overview

This is the Go implementation of the DevOps Info Service as a bonus task. It provides the same functionality as the Python version but compiled to a single binary.

## Implementation Details

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Returns service, system, runtime, and request information |
| `/health` | GET | Returns health status and uptime |

### Code Structure

```
app_go/
├── main.go             # Main application code
├── go.mod              # Go module file
├── README.md           # Documentation
└── docs/
    └── LAB01.md       
    └── GO.md        
    └──screenshots
```

### Key Features

- **Structs** — Used Go structs for type-safe JSON responses
- **Standard Library** — Only uses Go's built-in packages (no external dependencies)
- **Environment Variables** — Configurable port via `PORT` env variable
- **Error Handling** — Proper error handling for hostname and server startup

## Building and Running

### Development Mode

```bash
cd app_go
go run main.go
```

### Production Build

```bash
go build -o devops-info-service.exe main.go
.\devops-info-service.exe
```

### Custom Port

```powershell
$env:PORT=3000
go run main.go
```

## Testing

```bash
# Main endpoint
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health
```

## Comparison with Python Version

| Aspect | Python | Go |
|--------|--------|-----|
| Framework | FastAPI | net/http (standard library) |
| Dependencies | uvicorn, fastapi, psutil | None |
| Binary Size | ~50 MB (with venv) | ~8 MB |
| Startup Time | ~2 seconds | ~0.9 seconds |
| Runtime Required | Python interpreter | None |

## Challenges and Solutions

### Challenge: JSON Response Structure

**Problem:** Needed nested JSON structure matching the Python version.

**Solution:** Created multiple structs that reference each other:

```go
type HomeResponse struct {
    Service   ServiceInfo `json:"service"`
    System    SystemInfo  `json:"system"`
    Runtime   RuntimeInfo `json:"runtime"`
}

```

## What I Learned

1. Go's syntax is simpler than expected
2. Structs with JSON tags make API responses easy
3. Go's standard library is powerful — no frameworks needed
4. Compiled binaries are much smaller and faster than interpreted code
5. Go is widely used in DevOps tooling

## Conclusion

Building this service in Go was a great learning experience. The language is fun to work with, and I can see why tools like Kubernetes and Docker chose Go. The compiled binary is small, fast, and has no dependencies — perfect for containerized deployments.