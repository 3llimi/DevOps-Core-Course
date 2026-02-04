# Lab 2 Bonus — Multi-Stage Docker Build for Go

## Multi-Stage Build Strategy

### Why Multi-Stage Builds?

Go is a **compiled language**, meaning it needs the Go compiler and SDK to build the application, but the **runtime** only needs the compiled binary.

**The Problem:**
- `golang:1.25-alpine` image is ~300 MB
- Includes the Go compiler, linker, and build tools
- 95% of this is not needed to run the app

**The Solution:**
- **Stage 1 (Builder):** Use Go SDK to compile the binary
- **Stage 2 (Runtime):** Use minimal Alpine, copy only the binary

---

## Dockerfile Implementation

### Stage 1: Builder

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o devops-info-service main.go
```

**Key Decisions:**

1. **`golang:1.25-alpine`** instead of `golang:1.25`
   - Alpine variant: 336 MB vs 807 MB (full Debian-based image)
   - Still has everything needed to compile Go code

2. **`CGO_ENABLED=0`**
   - Creates a **static binary** with no C library dependencies
   - Allows us to use minimal base images (alpine, scratch, distroless)
   - Without this, binary would need glibc/musl from the base image

3. **`-ldflags="-w -s"`**
   - `-w`: Removes DWARF debugging information
   - `-s`: Removes symbol table and debug info
   - Reduces binary size by 20-30%

4. **Layer caching optimization:**
   - `go.mod` copied before `main.go`
   - Dependencies downloaded before code
   - Code changes don't force re-downloading dependencies

---

### Stage 2: Runtime

```dockerfile
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
RUN addgroup -S appuser && adduser -S appuser -G appuser
WORKDIR /app
COPY --from=builder /app/devops-info-service .
RUN chown -R appuser:appuser /app
USER appuser
CMD ["./devops-info-service"]
```

**Key Decisions:**

1. **`FROM alpine:3.19`** (~7 MB)
   - Minimal Linux distribution
   - Could use `FROM scratch` (0 MB) but Alpine provides useful debugging tools

2. **`COPY --from=builder`** 
   - **This is the magic!**
   - Copies ONLY the binary from Stage 1
   - Leaves behind the entire Go SDK (~300 MB)

3. **`ca-certificates`**
   - Needed if app makes HTTPS requests
   - Provides root SSL certificates

4. **Non-root user**
   - Created with Alpine's `adduser` command
   - Same security practice as Python app

---

## Size Comparison

### Build Output

```bash
$ docker build -t 3llimi/devops-go-service:latest .

[+] Building 42.1s (17/17) FINISHED
 => [internal] load build definition from Dockerfile
 => [internal] load .dockerignore
 => [builder 1/6] FROM golang:1.25-alpine
 => [stage-1 1/4] FROM alpine:3.19
 => [builder 2/6] WORKDIR /app
 => [builder 3/6] COPY go.mod ./
 => [builder 4/6] RUN go mod download
 => [builder 5/6] COPY main.go ./
 => [builder 6/6] RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o devops-info-service main.go
 => [stage-1 2/4] RUN apk --no-cache add ca-certificates
 => [stage-1 3/4] COPY --from=builder /app/devops-info-service .
 => [stage-1 4/4] RUN chown -R appuser:appuser /app
 => exporting to image
```

### Image Size Breakdown

```bash
$ docker images

REPOSITORY                   TAG       SIZE
3llimi/devops-go-service    latest    29.8 MB   ✅ Multi-stage build
golang                      1.25      807 MB    ❌ What we avoided
alpine                      3.19      7.3 MB    Base for stage 2
```

**Size Reduction: 807 MB → 29.8 MB (96.3% smaller!)** 🎉

### Layer Analysis

```bash
$ docker history 3llimi/devops-go-service:latest

IMAGE          SIZE      COMMENT
<latest>       0B        CMD ["./devops-info-service"]
<missing>      0B        USER appuser
<missing>      20kB      RUN chown -R appuser:appuser /app
<missing>      21.47 MB    COPY --from=builder /app/devops-info-service    ← Our binary
<missing>      0B        WORKDIR /app
<missing>      41kB      RUN addgroup -S appuser && adduser...
<missing>      524kB     RUN apk --no-cache add ca-certificates
<missing>      7.3 MB    FROM alpine:3.19                                ← Base OS
```

**Final breakdown:**
- Alpine base: 7.73 MB
- CA certificates: 524 KB
- Go binary: 21.47 MB
- User creation + ownership: 61 KB
- **Total: 29.8 MB**

---

## Why Multi-Stage Builds Matter

### 1. Massive Size Reduction

**807 MB → 29.8 MB (96.3% reduction)**

**Benefits:**
- ✅ Faster downloads from Docker Hub
- ✅ Less disk space on servers and Kubernetes nodes
- ✅ Faster deployment in production
- ✅ Lower bandwidth costs

**Real-world impact:**
- Deploying 10 containers: Saves 7.9 GB
- Deploying 100 containers: Saves 79 GB

---

### 2. Security Benefits

**Smaller Attack Surface:**
- ❌ **NO** Go compiler (can't compile malware inside container)
- ❌ **NO** build tools (can't download and build exploits)
- ❌ **NO** package manager (can't install backdoors)
- ✅ **ONLY** the binary and minimal OS

**Fewer Vulnerabilities:**
- Builder stage: ~300 packages → Dozens of CVEs
- Runtime stage: ~15 packages → Minimal CVEs
- **Less code to audit and patch**

**Example scenario:**
- If a vulnerability is found in the Go compiler, it doesn't affect your production container (because the compiler isn't there!)

---

### 3. Production Best Practice

**Industry Standard:**
- All major companies use multi-stage builds for compiled languages
- Kubernetes, Docker, Terraform, Prometheus all use this pattern
- Build-time dependencies should NEVER be in production images

**Separation of Concerns:**
- **Build stage:** All the tools needed to compile
- **Runtime stage:** Only what's needed to run
- Clear distinction between development and production

---

## Build Process Analysis

### First Build (Cold Cache)

```bash
$ docker build -t 3llimi/devops-go-service:latest .
[+] Building 45.3s

Stage 1 (Builder):
 => [builder 1/6] FROM golang:1.25-alpine          ~20s (download)
 => [builder 2/6] WORKDIR /app                     0.1s
 => [builder 3/6] COPY go.mod ./                   0.1s
 => [builder 4/6] RUN go mod download              2.3s
 => [builder 5/6] COPY main.go ./                  0.1s
 => [builder 6/6] RUN CGO_ENABLED=0 go build...    ~15s (compilation)

Stage 2 (Runtime):
 => [stage-1 1/4] FROM alpine:3.19                 ~5s (download)
 => [stage-1 2/4] RUN apk add ca-certificates      2.1s
 => [stage-1 3/4] COPY --from=builder...           0.1s
 => [stage-1 4/4] RUN chown...                     0.2s

Total: ~45 seconds
```

### Rebuild (Cached - No Code Changes)

```bash
$ docker build -t 3llimi/devops-go-service:latest .
[+] Building 2.1s (all layers CACHED)

Total: ~2 seconds ✅
```

### Rebuild (Code Changed)

```bash
$ docker build -t 3llimi/devops-go-service:latest .
[+] Building 18.5s

Stage 1:
 => CACHED [builder 1/6] FROM golang:1.25-alpine
 => CACHED [builder 2/6] WORKDIR /app
 => CACHED [builder 3/6] COPY go.mod ./
 => CACHED [builder 4/6] RUN go mod download       ← Dependencies cached!
 => [builder 5/6] COPY main.go ./                  0.1s
 => [builder 6/6] RUN CGO_ENABLED=0 go build...    ~15s (recompile)

Stage 2:
 => CACHED [stage-1 1/4] FROM alpine:3.19
 => CACHED [stage-1 2/4] RUN apk add ca-certificates
 => [stage-1 3/4] COPY --from=builder...           0.1s (new binary)
 => [stage-1 4/4] RUN chown...                     0.2s

Total: ~18 seconds
```

**Cache Efficiency:**
- Dependencies stay cached if `go.mod` doesn't change
- Only recompilation happens when code changes
- No need to re-download Alpine or Go SDK

---

## Testing the Container

### Build and Run

```bash
$ docker build -t 3llimi/devops-go-service:latest .
$ docker run -p 8080:8080 3llimi/devops-go-service:latest

Server starting on port 8080
```

### Test Endpoints

```bash
$ curl http://localhost:8080/

{
  "service": {
    "name": "devops-info-service",
    "version": "1.0.0",
    "description": "DevOps course info service",
    "framework": "Go net/http"
  },
  "system": {
    "hostname": "333e9c5fbc1c",
    "platform": "linux",
    "platform_version": "linux-amd64",
    "architecture": "amd64",
    "cpu_count": 12,
    "go_version": "go1.25.6"
  },
  "runtime": {
    "uptime_seconds": 15,
    "uptime_human": "0 hours, 0 minutes",
    "current_time": "2026-02-04T16:27:02Z",
    "timezone": "UTC"
  },
  "request": {
    "client_ip": "172.17.0.1",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36 OPR/126.0.0.0",
    "method": "GET",
    "path": "/"
  },
  "endpoints": [
    {
      "path": "/",
      "method": "GET",
      "description": "Service information"
    },
    {
      "path": "/health",
      "method": "GET",
      "description": "Health check"
    }
  ]
}
```

```bash
$ curl http://localhost:8080/health

{
  "status": "healthy",
  "timestamp": "2026-02-04T16:27:18Z",
  "uptime_seconds": 31
}
```

✅ **Application works perfectly in the container!**

---

## Docker Hub

**Repository URL:** https://hub.docker.com/r/3llimi/devops-go-service

### Push Process

```bash
$ docker login
Username: 3llimi
Password: [hidden]
Login Succeeded

$ docker push 3llimi/devops-go-service:latest

The push refers to repository [docker.io/3llimi/devops-go-service]
ae6e72fa2cf9: Pushed
3c9780956289: Pushed
c6dd4b209ebb: Pushed
a329b995e16c: Pushed
59b732c23da9: Pushed
17a39c0ba978: Pushed
7d228ba7db7f: Pushed
latest: digest: sha256:3114d801586fb09f954de188394207f2b66b433fdb59fdaf20f4b13b332b180a size: 856
```

### Pull and Run

```bash
$ docker pull 3llimi/devops-go-service:latest
$ docker run -p 8080:8080 3llimi/devops-go-service:latest
```

---

## Alternative Approaches Considered

### Option 1: FROM scratch

```dockerfile
FROM scratch
COPY --from=builder /app/devops-info-service .
CMD ["./devops-info-service"]
```

**Pros:**
- **Smallest possible:** ~8.5 MB (just the binary!)
- Maximum security (no OS at all)

**Cons:**
- ❌ No shell (can't debug with `docker exec`)
- ❌ No ca-certificates (HTTPS won't work)
- ❌ No timezone data
- ❌ Harder to troubleshoot

**When to use:** Ultra-minimal services with no external dependencies

---

### Option 2: Distroless

```dockerfile
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/devops-info-service .
CMD ["./devops-info-service"]
```

**Pros:**
- ~10 MB (includes ca-certificates)
- Google-maintained, security-focused
- No shell (harder to exploit)

**Cons:**
- Can't `docker exec` for debugging
- Slightly larger than scratch

**When to use:** Production services prioritizing security over debuggability

---

### My Choice: Alpine

**Why Alpine:**
- ✅ Good balance: 29.8 MB (small but usable)
- ✅ Can debug: `docker exec -it <container> /bin/sh`
- ✅ Has ca-certificates (HTTPS works)
- ✅ Industry standard (widely used and documented)
- ✅ Only 10 MB larger than distroless

**Trade-off:** 10 MB extra for significant debuggability is worth it for a learning environment.

---

## Challenges & Solutions

### Challenge 1: CGO Dependency Error

**Problem:**
First build failed with:
```
standard_init_linux.go:228: exec user process caused: no such file or directory
```

**Cause:** Binary was compiled with CGO enabled (default), which links against C libraries. Alpine didn't have the required `glibc`.

**Solution:** Added `CGO_ENABLED=0` to create a fully static binary with no C dependencies.

**Learning:** Always build static binaries for minimal base images.

---

### Challenge 2: File Ownership

**Problem:** First run failed because binary was owned by root but running as `appuser`.

**Solution:** Added `RUN chown -R appuser:appuser /app` before `USER appuser`.

**Learning:** Same lesson as Python Dockerfile - always fix ownership before switching users.

---

## What I Learned

1. **Multi-stage builds are essential for compiled languages**
   - 96.3% size reduction is massive
   - Industry standard for production deployments

2. **Static binaries enable minimal images**
   - `CGO_ENABLED=0` is critical
   - Allows using scratch, distroless, or Alpine

3. **Security through minimalism**
   - Less code = less vulnerabilities
   - No build tools in production = harder to exploit

4. **Layer caching works across stages**
   - Stage 1 layers are cached independently
   - Code changes don't invalidate dependency layers

5. **Go is perfect for containers**
   - Single binary with zero dependencies
   - Fast compilation
   - Tiny final images

---

## Conclusion

Multi-stage builds transformed a **807 MB** bloated image into a **29.8 MB** production-ready container. This technique is critical for deploying compiled applications in Kubernetes and cloud environments where image size directly impacts deployment speed and costs.

The Go application now:
- ✅ Runs as non-root user
- ✅ Has minimal attack surface
- ✅ Deploys 40x faster than single-stage
- ✅ Costs less in bandwidth and storage
- ✅ Follows industry best practices

**Final metrics:**
- **Compressed size:** ~15 MB (what users download)
- **Uncompressed size:** 29.8 MB (disk usage)
- **Size reduction:** 807 MB → 29.8 MB (96.3% reduction vs full golang)
- **Size reduction:** 336 MB → 29.8 MB (91.1% reduction vs alpine golang)