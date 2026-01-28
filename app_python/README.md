# DevOps Info Service

A Python web service that provides system and runtime information. Built with FastAPI for the DevOps Core Course.

## Overview

This service exposes REST API endpoints that return:
- Service metadata (name, version, framework)
- System information (hostname, platform, CPU, Python version)
- Runtime information (uptime, current time)
- Request details (client IP, user agent)

## Prerequisites

- Python 3.11 or higher
- pip (Python package manager)

## Installation

```bash
# Navigate to app folder
cd app_python

# Create virtual environment
python -m venv venv

# Activate virtual environment (Windows PowerShell)
.\venv\Scripts\Activate

# Activate virtual environment (Linux/Mac)
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt
```

## Running the Application

**Default (port 8000):**
```bash
python app.py
```

**Custom port:**
```bash
# Windows PowerShell
$env:PORT=3000
python app.py

# Linux/Mac
PORT=3000 python app.py
```

**Custom host and port:**
```bash
# Windows PowerShell
$env:HOST="127.0.0.1"
$env:PORT=5000
python app.py
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service and system information |
| `/health` | GET | Health check for monitoring |
| `/docs` | GET | Swagger UI documentation |

### GET `/` — Main Endpoint

Returns comprehensive service and system information.

**Request:**
```bash
curl http://localhost:8000/
```

**Response:**
```json
{
  "service": {
    "name": "devops-info-service",
    "version": "1.0.0",
    "description": "DevOps course info service",
    "framework": "FastAPI"
  },
  "system": {
    "hostname": "3llimi",
    "platform": "Windows",
    "platform_version": "Windows-11-10.0.26200-SP0",
    "architecture": "AMD64",
    "cpu_count": 12,
    "python_version": "3.14.2"
  },
  "runtime": {
    "uptime_seconds": 58,
    "uptime_human": "0 hours, 0 minutes",
    "current_time": "2026-01-26T18:54:58.321970+00:00",
    "timezone": "UTC"
  },
  "request": {
    "client_ip": "127.0.0.1",
    "user_agent": "curl/7.81.0",
    "method": "GET",
    "path": "/"
  },
  "endpoints": [
    {"path": "/", "method": "GET", "description": "Service information"},
    {"path": "/health", "method": "GET", "description": "Health check"}
  ]
}
```

### GET `/health` — Health Check

Returns service health status for monitoring and Kubernetes probes.

**Request:**
```bash
curl http://localhost:8000/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-26T18:55:51.887474+00:00",
  "uptime_seconds": 51
}
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `HOST` | `0.0.0.0` | Server bind address |
| `PORT` | `8000` | Server port |

## Project Structure

```
app_python/
├── app.py              # Main application
├── requirements.txt    # Dependencies
├── .gitignore         # Git ignore rules
├── README.md          # This file
├── tests/             # Unit tests
│   └── __init__.py
└── docs/
    ├── LAB01.md       # Lab submission
    └── screenshots/   # Proof of work
```

## Tech Stack

- **Language:** Python 3.14
- **Framework:** FastAPI 0.115.0
- **Server:** Uvicorn 0.32.0