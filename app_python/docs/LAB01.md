# Lab 1 — DevOps Info Service: Submission

## Framework Selection

### My Choice: FastAPI

I chose **FastAPI** for building this DevOps info service.

### Comparison with Alternatives

FastAPI is a good choice for APIs because it’s fast, supports async, and automatically generates API documentation, and it’s becoming more popular in the tech industry with growing demand in job listings. Even though Flask is easier and good for small projects, but it’s slower, synchronous, and needs manual documentation. Django is better for full web applications, widely used in companies with larger projects, but it has a steeper learning curve and can feel heavy for simple use cases.

### Why I Chose FastAPI

1. **Automatic API Documentation** — Swagger UI is generated automatically at `/docs`, which makes testing and sharing the API easy.

2. **Modern Python** — FastAPI uses type hints and async/await, which are modern Python features that are good to learn.

3. **Great for Microservices** — FastAPI is lightweight and fast, perfect for the DevOps info service we're building.

4. **Performance** — Built on Starlette and Pydantic, FastAPI is one of the fastest Python frameworks.

### Why Not Flask

Flask is simpler but doesn't have built-in documentation or type validation. Would need extra libraries.

### Why Not Django

Django is too heavy for a simple API service. It includes ORM, admin panel, and templates that we don't need.

---

## Best Practices Applied

### 1. Clean Code Organization

Imports are grouped properly:
```python
# Standard library
from datetime import datetime, timezone
import platform
import socket
import os

# Third-party
from fastapi import FastAPI, Request
```

### 2. Configuration via Environment Variables

```python
HOST = os.getenv('HOST', '0.0.0.0')
PORT = int(os.getenv('PORT', 8000))
```

**Why it matters:** Allows changing configuration without modifying code. Essential for Docker and Kubernetes deployments.

### 3. Helper Functions

```python
def get_uptime():
    delta = datetime.now(timezone.utc) - START_TIME
    secs = int(delta.total_seconds())
    hrs = secs // 3600
    mins = (secs % 3600) // 60
    return {
        "seconds": secs,
        "human": f"{hrs} hours, {mins} minutes"
    }
```

**Why it matters:** Reusable code — used in both `/` and `/health` endpoints.

### 4. Consistent JSON Responses

All endpoints return structured JSON with consistent formatting.

### 5. Safe Defaults

```python
"client_ip": request.client.host if request.client else "unknown"
```

**Why it matters:** Prevents crashes if a value is missing.

---

## API Documentation

### Endpoint: GET `/`

**Description:** Returns service and system information.

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
    "architecture": "AMD64",
    "cpu_count": 12,
    "python_version": "3.14.2"
  },
  "runtime": {
    "uptime_seconds": 58,
    "uptime_human": "0 hours, 0 minutes",
    "current_time": "2026-01-26T18:54:58+00:00",
    "timezone": "UTC"
  },
  "request": {
    "client_ip": "127.0.0.1",
    "user_agent": "Mozilla/5.0...",
    "method": "GET",
    "path": "/"
  },
  "endpoints": [...]
}
```

### Endpoint: GET `/health`

**Description:** Health check for monitoring and Kubernetes probes.

**Request:**
```bash
curl http://localhost:8000/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-26T18:55:51+00:00",
  "uptime_seconds": 51
}
```

---

## Testing Evidence

### Testing Commands Used

```bash
# Start the application
python app.py

# Test main endpoint
curl http://localhost:8000/

# Test health endpoint
curl http://localhost:8000/health

# Test with custom port
$env:PORT=3000
python app.py
curl http://localhost:3000/

# View Swagger documentation
# Open http://localhost:8000/docs in browser
```

### Screenshots

1. **01-main-endpoint.png** — Main endpoint showing complete JSON response
2. **02-health-check.png** — Health check endpoint response
3. **03-formatted-output.png** — Swagger UI documentation

---

## Challenges & Solutions

### Challenge 1: Understanding Request Object

**Problem:** Wasn't sure how to get client IP and user agent in FastAPI.

**Solution:** Import `Request` from FastAPI and add it as a parameter:
```python
from fastapi import FastAPI, Request

@app.get("/")
def home(request: Request):
    client_ip = request.client.host
    user_agent = request.headers.get("user-agent")
```

### Challenge 2: Timezone-Aware Timestamps

**Problem:** Needed UTC timestamps for consistency across different servers.

**Solution:** Used `timezone.utc` from datetime module:
```python
from datetime import datetime, timezone

current_time = datetime.now(timezone.utc).isoformat()
```

### Challenge 3: Running with Custom Port

**Problem:** Needed to make the port configurable.

**Solution:** Used environment variables with a default value:
```python
import os
PORT = int(os.getenv('PORT', 8000))
```

---

## GitHub Community

### Why Starring Repositories Matters

Starring repositories is important in open source because it:
- Bookmarks useful projects for later reference
- Shows appreciation to maintainers
- Helps projects gain visibility and attract contributors
- Indicates project quality to other developers

### How Following Developers Helps

Following developers on GitHub helps in team projects and professional growth by:
- Keeping you updated on teammates' and mentors' activities
- Discovering new projects through their activity
- Learning from experienced developers' code and commits
- Building professional connections in the developer community

### Completed Actions

- [x] Starred course repository
- [x] Starred [simple-container-com/api](https://github.com/simple-container-com/api)
- [x] Followed [@Cre-eD](https://github.com/Cre-eD)
- [x] Followed [@marat-biriushev](https://github.com/marat-biriushev)
- [x] Followed [@pierrepicaud](https://github.com/pierrepicaud)
- [x] Followed 3 classmates [@abdughafforzoda](https://github.com/abdughafforzoda),[@Boogyy](https://github.com/Boogyy), [@mpasgat](https://github.com/mpasgat)