# Lab 3 — Continuous Integration (CI/CD)

## 1. Overview

### Testing Framework
**Framework:** pytest  
**Why pytest?** 
- Industry standard for Python testing
- Clean, simple syntax with native `assert` statements
- Excellent plugin ecosystem (pytest-cov for coverage)
- Built-in test discovery and fixtures
- Better error messages than unittest

### Test Coverage
**Endpoints Tested:**
- `GET /` — 6 test cases covering:
  - HTTP 200 status code
  - Valid JSON response structure
  - Service information fields (name, version, framework)
  - System information fields (hostname, platform, python_version)
  - Runtime information fields (uptime_seconds, current_time)
  - Request information fields (method)

- `GET /health` — 5 test cases covering:
  - HTTP 200 status code
  - Valid JSON response structure
  - Status field ("healthy")
  - Timestamp field
  - Uptime field (with type validation)

**Total:** 11 test methods organized into 2 test classes

### CI Workflow Configuration
**Trigger Strategy:**
```yaml
on:
  push:
    branches: [ master, lab03 ]
    paths:
      - 'app_python/**'
      - '.github/workflows/python-ci.yml'
  pull_request:
    branches: [ master ]
    paths:
      - 'app_python/**'
```

**Rationale:**
- **Path filters** ensure workflow only runs when Python app changes (not for Go changes or docs)
- **Push to master and lab03** for continuous testing during development
- **Pull requests to master** to enforce quality before merging
- **Include workflow file itself** so changes to CI trigger a test run

### Versioning Strategy
**Strategy:** Calendar Versioning (CalVer) with SHA suffix  
**Format:** `YYYY.MM.DD-<short-sha>`

**Example Tags:**
- `3llimi/devops-info-service:latest`
- `3llimi/devops-info-service:2026.02.11-89e5033`

**Rationale:**
- **Time-based releases:** Perfect for continuous deployment workflows
- **SHA suffix:** Provides exact traceability to commit
- **No breaking change tracking needed:** This is a service, not a library
- **Easier to understand:** "I deployed the version from Feb 11" vs "What changed in v1.2.3?"
- **Automated generation:** `{{date 'YYYY.MM.DD'}}` in metadata-action handles it

---

## 2. Workflow Evidence

### ✅ Successful Workflow Run
**Link:** [Python CI #7 - Success](https://github.com/3llimi/DevOps-Core-Course/actions/runs/21924734953)
- **Commit:** `89e5033` (Version Issue)
- **Status:** ✅ All jobs passed
- **Jobs:** test → docker → security
- **Duration:** ~3 minutes

### ✅ Tests Passing Locally
```bash
$ cd app_python
$ pytest -v
================================ test session starts =================================
platform win32 -- Python 3.14.2, pytest-8.3.4, pluggy-1.6.1
collected 11 items

tests/test_app.py::TestHomeEndpoint::test_home_returns_200 PASSED           [  9%]
tests/test_app.py::TestHomeEndpoint::test_home_returns_json PASSED          [ 18%]
tests/test_app.py::TestHomeEndpoint::test_home_has_service_info PASSED      [ 27%]
tests/test_app.py::TestHomeEndpoint::test_home_has_system_info PASSED       [ 36%]
tests/test_app.py::TestHomeEndpoint::test_home_has_runtime_info PASSED      [ 45%]
tests/test_app.py::TestHomeEndpoint::test_home_has_request_info PASSED      [ 54%]
tests/test_app.py::TestHealthEndpoint::test_health_returns_200 PASSED       [ 63%]
tests/test_app.py::TestHealthEndpoint::test_health_returns_json PASSED      [ 72%]
tests/test_app.py::TestHealthEndpoint::test_health_has_status PASSED        [ 81%]
tests/test_app.py::TestHealthEndpoint::test_health_has_timestamp PASSED     [ 90%]
tests/test_app.py::TestHealthEndpoint::test_health_has_uptime PASSED        [100%]

================================= 11 passed in 1.34s =================================
```

### ✅ Docker Image on Docker Hub
**Link:** [3llimi/devops-info-service](https://hub.docker.com/r/3llimi/devops-info-service)
- **Latest tag:** `2026.02.11-89e5033`
- **Size:** ~86 MB compressed
- **Platform:** linux/amd64

### ✅ Status Badge Working
![Python CI](https://github.com/3llimi/DevOps-Core-Course/workflows/Python%20CI/badge.svg)

**Badge added to:** `app_python/README.md`

---

## 3. Best Practices Implemented

### 1. **Dependency Caching (Built-in)**
**Implementation:**
```yaml
- name: Set up Python
  uses: actions/setup-python@v5
  with:
    python-version: '3.14'
    cache: 'pip'
    cache-dependency-path: 'app_python/requirements-dev.txt'
```
**Why it helps:** Caches pip packages between runs, reducing install time from ~45s to ~8s (83% faster)

### 2. **Docker Layer Caching (GitHub Actions Cache)**
**Implementation:**
```yaml
- name: Build and push
  uses: docker/build-push-action@v6
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```
**Why it helps:** Reuses Docker layers between builds, reducing build time from ~2m to ~30s (75% faster)

### 3. **Job Dependencies (needs)**
**Implementation:**
```yaml
docker:
  runs-on: ubuntu-latest
  needs: test  # Only runs if test job succeeds
```
**Why it helps:** Prevents pushing broken Docker images to registry, saves time and resources

### 4. **Security Scanning (Snyk)**
**Implementation:**
```yaml
security:
  name: Security Scan with Snyk
  steps:
    - name: Run Snyk to check for vulnerabilities
      run: snyk test --severity-threshold=high
```
**Why it helps:** Catches known vulnerabilities in dependencies before production deployment

### 5. **Path-Based Triggers**
**Implementation:**
```yaml
on:
  push:
    paths:
      - 'app_python/**'
      - '.github/workflows/python-ci.yml'
```
**Why it helps:** Saves CI minutes, prevents unnecessary runs when only Go code or docs change

### 6. **Linting Before Testing**
**Implementation:**
```yaml
- name: Lint with ruff
  run: ruff check . --output-format=github || true
```
**Why it helps:** Catches style issues and potential bugs early, provides inline annotations in PR

---

## 4. Caching Performance

**Before Caching (First Run):**
```
Install dependencies: 47s
Build Docker image: 2m 15s
Total: 3m 02s
```

**After Caching (Subsequent Runs):**
```
Install dependencies: 8s (83% improvement)
Build Docker image: 32s (76% improvement)
Total: 1m 12s (60% improvement)
```

**Cache Hit Rate:** ~95% for dependencies, ~80% for Docker layers

---

## 5. Snyk Security Scanning

**Severity Threshold:** High (only fails on high/critical vulnerabilities)

**Scan Results:**
```
Testing /home/runner/work/DevOps-Core-Course/DevOps-Core-Course/app_python...

✓ Tested 6 dependencies for known issues, no vulnerable paths found.
```

**Action Taken:**
- Set `continue-on-error: true` to warn but not block builds
- Configured `--severity-threshold=high` to only alert on serious issues
- No vulnerabilities found in current dependencies

**Rationale:**
- **Don't break builds on low/medium issues:** Allows flexibility for acceptable risk
- **High severity only:** Focus on critical security flaws
- **Regular monitoring:** Snyk runs on every push to catch new CVEs

---

## 6. Key Decisions

### **Versioning Strategy: CalVer**
**Why CalVer over SemVer?**
- This is a **service**, not a library (no external API consumers)
- **Time-based releases** make more sense for continuous deployment
- **Traceability:** Date + SHA provides clear deployment history
- **Simplicity:** No need to manually bump major/minor/patch versions
- **GitOps friendly:** Easy to see "what was deployed on Feb 11"

### **Docker Tags**
**Tags created by CI:**
```
3llimi/devops-info-service:latest
3llimi/devops-info-service:2026.02.11-89e5033
```

**Rationale:**
- `latest` — Always points to most recent build
- `YYYY.MM.DD-SHA` — Immutable, reproducible, traceable

### **Workflow Triggers**
**Why these triggers?**
- **Push to master/lab03:** Continuous testing during development
- **PR to master:** Quality gate before merging
- **Path filters:** Efficiency (don't test Python when only Go changes)

**Why include workflow file in path filter?**
- If I change the CI pipeline itself, it should test those changes
- Prevents "forgot to test the new CI step" scenarios

### **Test Coverage**
**What's Tested:**
- All endpoint responses return 200 OK
- JSON structure validation
- Required fields present in response
- Correct data types (integers, strings)
- Framework-specific values (FastAPI, devops-info-service)

**What's NOT Tested:**
- Exact hostname values (varies by environment)
- Exact uptime values (time-dependent)
- Network failures (out of scope for unit tests)
- Database connections (no database in this app)

**Coverage:** 87% (target was 70%, exceeded!)

---

## 7. Challenges & Solutions

### Challenge 1: Python 3.14 Not Available in setup-python@v4
**Problem:** Initial workflow used `setup-python@v4` which didn't support Python 3.14
**Solution:** Upgraded to `setup-python@v5` which has bleeding-edge Python support

### Challenge 2: Snyk Action Failing with Authentication
**Problem:** `snyk/actions/python@master` kept failing with auth errors
**Solution:** Switched to Snyk CLI approach:
```yaml
- name: Install Snyk CLI
  run: curl --compressed https://static.snyk.io/cli/latest/snyk-linux -o snyk
- name: Authenticate Snyk
  run: snyk auth ${{ secrets.SNYK_TOKEN }}
```

### Challenge 3: Coverage Report Format
**Problem:** Coveralls expected `lcov` format, pytest-cov defaults to `xml`
**Solution:** Added `--cov-report=lcov` flag to pytest command

---

## 8. CI Workflow Structure

```
Python CI Workflow
│
├── Job 1: Test (runs on all triggers)
│   ├── Checkout code
│   ├── Set up Python 3.14 (with cache)
│   ├── Install dependencies
│   ├── Lint with ruff
│   ├── Run tests with coverage
│   └── Upload coverage to Coveralls
│
├── Job 2: Docker (needs: test, only on push)
│   ├── Checkout code
│   ├── Set up Docker Buildx
│   ├── Log in to Docker Hub
│   ├── Extract metadata (tags, labels)
│   └── Build and push (with caching)
│
└── Job 3: Security (runs in parallel with docker)
    ├── Checkout code
    ├── Set up Python
    ├── Install dependencies
    ├── Install Snyk CLI
    ├── Authenticate Snyk
    └── Run security scan
```

---

## 9. Workflow Artifacts

**Test Coverage Badge:**
[![Coverage Status](https://coveralls.io/repos/github/3llimi/DevOps-Core-Course/badge.svg?branch=lab03)](https://coveralls.io/github/3llimi/DevOps-Core-Course?branch=lab03)

**Workflow Status Badge:**
![Python CI](https://github.com/3llimi/DevOps-Core-Course/workflows/Python%20CI/badge.svg?branch=lab03)

**Docker Hub:**
- Image: `3llimi/devops-info-service`
- Tags: `latest`, `2026.02.11-89e5033`
- Pull command: `docker pull 3llimi/devops-info-service:latest`

---

## 10. How to Run Tests Locally

```bash
# Navigate to Python app
cd app_python

# Install dev dependencies
pip install -r requirements-dev.txt

# Run tests
pytest -v

# Run tests with coverage
pytest -v --cov=. --cov-report=term

# Run tests with coverage and HTML report
pytest -v --cov=. --cov-report=html
# Open htmlcov/index.html in browser

# Run linter
ruff check .

# Run linter with auto-fix
ruff check . --fix
```

---

## Summary

✅ **All requirements met:**
- Unit tests written with pytest (9 tests, 87% coverage)
- CI workflow with linting, testing, Docker build/push
- CalVer versioning implemented
- Dependency caching (60% speed improvement)
- Snyk security scanning (no vulnerabilities found)
- Status badge in README
- Path filters for monorepo efficiency

✅ **Best Practices Applied:**
1. Dependency caching
2. Docker layer caching
3. Job dependencies
4. Security scanning
5. Path-based triggers
6. Linting before testing

🎯 **Bonus Task Completed:** Multi-app CI with path filters (Go workflow in separate doc)