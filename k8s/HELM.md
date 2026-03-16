# Lab 10 — Helm Package Manager

## Chart Structure

```
k8s/
├── common-lib/                          # Library chart (shared templates)
│   ├── Chart.yaml                       # type: library
│   └── templates/
│       └── _helpers.tpl                 # Shared: labels, fullname, selectorLabels
│
├── devops-python/                       # Python app chart
│   ├── Chart.yaml                       # Declares common-lib dependency
│   ├── values.yaml                      # Default values
│   ├── values-dev.yaml                  # Dev environment overrides
│   ├── values-prod.yaml                 # Prod environment overrides
│   ├── charts/
│   │   └── common-lib-0.1.0.tgz        # Packaged dependency
│   └── templates/
│       ├── deployment.yaml              # Uses common.* templates
│       ├── service.yaml                 # Uses common.* templates
│       ├── _helpers.tpl                 # Chart-specific helpers (kept for reference)
│       ├── NOTES.txt                    # Post-install instructions
│       └── hooks/
│           ├── pre-install-job.yaml     # Runs before install
│           └── post-install-job.yaml    # Runs after install
│
└── devops-go/                           # Go app chart
    ├── Chart.yaml                       # Declares common-lib dependency
    ├── values.yaml                      # Default values
    ├── charts/
    │   └── common-lib-0.1.0.tgz        # Packaged dependency
    └── templates/
        ├── deployment.yaml              # Uses common.* templates
        ├── service.yaml                 # Uses common.* templates
        ├── _helpers.tpl                 # Chart-specific helpers
        └── NOTES.txt                    # Post-install instructions
```

### Key Template Files

**`common-lib/templates/_helpers.tpl`** — Shared template library used by both app charts. Defines `common.name`, `common.fullname`, `common.chart`, `common.labels`, and `common.selectorLabels`. Eliminates duplication across charts.

**`devops-python/templates/deployment.yaml`** — Deployment template with configurable replicas, image, resources, and probes all driven from values. Rolling update strategy hardcoded as a production best practice.

**`devops-python/templates/service.yaml`** — Service template with conditional NodePort block — only renders `nodePort` field when service type is NodePort, making it compatible with ClusterIP too.

**`devops-python/templates/hooks/`** — Pre and post install Jobs that run lifecycle validation tasks. Use `before-hook-creation` deletion policy so they remain visible for inspection after execution.

---

## Configuration Guide

### Default `values.yaml` Structure

```yaml
replicaCount: 3

image:
  repository: 3llimi/devops-info-service
  tag: "latest"
  pullPolicy: IfNotPresent

service:
  type: NodePort
  port: 80
  targetPort: 8000
  nodePort: 30080

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

livenessProbe:
  httpGet:
    path: /health
    port: 8000
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8000
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3
```

### Key Values and Purpose

| Value | Purpose | Default |
|-------|---------|---------|
| `replicaCount` | Number of pod replicas | `3` |
| `image.repository` | Docker Hub image name | `3llimi/devops-info-service` |
| `image.tag` | Image tag — pin for production | `latest` |
| `image.pullPolicy` | When to pull image | `IfNotPresent` |
| `service.type` | NodePort or ClusterIP | `NodePort` |
| `service.targetPort` | Container port | `8000` |
| `resources.requests` | Scheduler placement hint | 100m CPU, 128Mi RAM |
| `resources.limits` | Hard resource ceiling | 200m CPU, 256Mi RAM |
| `livenessProbe` | Restart trigger config | /health, 10s delay |
| `readinessProbe` | Traffic readiness config | /health, 5s delay |

### Environment-Specific Values

**`values-dev.yaml`** — Minimal resources for local development:
```yaml
replicaCount: 1
image:
  tag: "latest"
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 50m
    memory: 64Mi
livenessProbe:
  initialDelaySeconds: 5
  periodSeconds: 10
```

**`values-prod.yaml`** — Full resources with pinned image tag:
```yaml
replicaCount: 5
image:
  tag: "2026.02.11-89e5033"
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 200m
    memory: 256Mi
livenessProbe:
  initialDelaySeconds: 30
  periodSeconds: 5
```

### Installation Commands

```bash
# Default install
helm install devops-python k8s/devops-python

# Development environment
helm install devops-python k8s/devops-python -f k8s/devops-python/values-dev.yaml

# Production environment
helm install devops-python k8s/devops-python -f k8s/devops-python/values-prod.yaml

# Override a single value
helm install devops-python k8s/devops-python --set replicaCount=2

# Upgrade existing release
helm upgrade devops-python k8s/devops-python -f k8s/devops-python/values-prod.yaml

# Rollback to previous revision
helm rollback devops-python 1

# Uninstall
helm uninstall devops-python
```

---

## Hook Implementation

### Hooks Overview

| Hook | Weight | Policy | Purpose |
|------|--------|--------|---------|
| `pre-install` | -5 | `before-hook-creation` | Environment validation before deployment |
| `post-install` | +5 | `before-hook-creation` | Smoke test after deployment |

**Why weight -5 and +5:** Lower weight runs first. Pre-install at -5 is guaranteed to complete before post-install at +5 starts.

**Why `before-hook-creation` policy:** Keeps completed job resources visible for inspection and logs. `hook-succeeded` deletes them immediately — harder to debug. In production, switch to `hook-succeeded` to keep the cluster clean.

### Pre-Install Hook

Runs before any chart resources are created. Simulates environment validation — in production this would check database connectivity, required secrets, or external service availability.

```yaml
annotations:
  "helm.sh/hook": pre-install
  "helm.sh/hook-weight": "-5"
  "helm.sh/hook-delete-policy": before-hook-creation
```

### Post-Install Hook

Runs after all chart resources are installed and ready. Simulates smoke testing — in production this would run HTTP health checks, integration tests, or send deployment notifications.

```yaml
annotations:
  "helm.sh/hook": post-install
  "helm.sh/hook-weight": "5"
  "helm.sh/hook-delete-policy": before-hook-creation
```

### Hook Execution Evidence

```bash
$ kubectl get jobs
NAME                                       STATUS     COMPLETIONS   DURATION   AGE
devops-python-devops-python-post-install   Complete   1/1           11s        26s
devops-python-devops-python-pre-install    Complete   1/1           10s        36s
```

```bash
$ kubectl describe job devops-python-devops-python-pre-install

Name:             devops-python-devops-python-pre-install
Annotations:      helm.sh/hook: pre-install
                  helm.sh/hook-delete-policy: before-hook-creation
                  helm.sh/hook-weight: -5
Start Time:       Mon, 16 Mar 2026 04:27:12 +0300
Completed At:     Mon, 16 Mar 2026 04:27:22 +0300
Duration:         10s
Pods Statuses:    0 Active (0 Ready) / 1 Succeeded / 0 Failed
Events:
  Normal  SuccessfulCreate  40s  job-controller  Created pod: devops-python-devops-python-pre-install-w8qpm
  Normal  Completed         30s  job-controller  Job completed
```

```bash
$ kubectl logs job/devops-python-devops-python-pre-install
Pre-install validation started
Checking environment...
Pre-install validation completed successfully

$ kubectl logs job/devops-python-devops-python-post-install
Post-install smoke test started
Verifying deployment...
Smoke test passed successfully
```

---

## Installation Evidence

### Helm List

```bash
$ helm list
NAME            NAMESPACE  REVISION  UPDATED                                STATUS    CHART               APP VERSION
devops-go       default    1         2026-03-16 04:34:16.8499444 +0300 MSK  deployed  devops-go-0.1.0     1.0.0
devops-python   default    1         2026-03-16 04:33:54.8292463 +0300 MSK  deployed  devops-python-0.1.0 1.0.0
```

### All Resources

```bash
$ kubectl get pods
NAME                                             READY   STATUS      RESTARTS   AGE
devops-go-devops-go-5f67859b64-9jqv6             1/1     Running     0          29s
devops-go-devops-go-5f67859b64-gm9cc             1/1     Running     0          29s
devops-go-devops-go-5f67859b64-nhw29             1/1     Running     0          29s
devops-python-devops-python-654d887bd9-5qm8p     1/1     Running     0          41s
devops-python-devops-python-654d887bd9-5xzm7     1/1     Running     0          41s
devops-python-devops-python-654d887bd9-zdjg7     1/1     Running     0          41s
devops-python-devops-python-post-install-8jtlr   0/1     Completed   0          41s
devops-python-devops-python-pre-install-6m9l8    0/1     Completed   0          52s

$ kubectl get services
NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
devops-go-devops-go           NodePort    10.100.26.230    <none>        80:30081/TCP   33s
devops-python-devops-python   NodePort    10.111.113.146   <none>        80:30080/TCP   45s
kubernetes                    ClusterIP   10.96.0.1        <none>        443/TCP        23m
```

### Multi-Environment Demonstration

**Dev environment (1 replica, minimal resources):**
```bash
$ helm upgrade devops-python k8s/devops-python -f k8s/devops-python/values-dev.yaml
Release "devops-python" has been upgraded. Happy Helming!
REVISION: 2 | Replicas: 1 | Image: latest
```

**Prod environment (5 replicas, pinned tag, full resources):**
```bash
$ helm upgrade devops-python k8s/devops-python -f k8s/devops-python/values-prod.yaml
Release "devops-python" has been upgraded. Happy Helming!
REVISION: 3 | Replicas: 5 | Image: 2026.02.11-89e5033
```

---

## Operations

### Deploy

```bash
# Install dependencies first
helm dependency update k8s/devops-python
helm dependency update k8s/devops-go

# Install both charts
helm install devops-python k8s/devops-python
helm install devops-go k8s/devops-go
```

### Upgrade

```bash
# Upgrade with new values
helm upgrade devops-python k8s/devops-python -f k8s/devops-python/values-prod.yaml

# Watch rollout
kubectl rollout status deployment/devops-python-devops-python
```

### Rollback

```bash
# View history
helm history devops-python

# Rollback to specific revision
helm rollback devops-python 1
```

### Uninstall

```bash
helm uninstall devops-python
helm uninstall devops-go
```

---

## Testing & Validation

### Lint

```bash
$ helm lint k8s/devops-python
==> Linting k8s/devops-python
[INFO] Chart.yaml: icon is recommended
1 chart(s) linted, 0 chart(s) failed

$ helm lint k8s/devops-go
==> Linting k8s/devops-go
[INFO] Chart.yaml: icon is recommended
1 chart(s) linted, 0 chart(s) failed
```

### Template Rendering

```bash
$ helm template devops-python k8s/devops-python
---
# Source: devops-python/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-python-devops-python
  labels:
    helm.sh/chart: devops-python-0.1.0
    app.kubernetes.io/name: devops-python
    app.kubernetes.io/instance: devops-python
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/managed-by: Helm
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: devops-python
    app.kubernetes.io/instance: devops-python
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8000
      nodePort: 30080
---
# Source: devops-python/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-python-devops-python
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: devops-python
        image: "3llimi/devops-info-service:latest"
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 100m
            memory: 128Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Dry Run

```bash
$ helm install --dry-run --debug devops-python k8s/devops-python

NAME: devops-python
STATUS: pending-install
REVISION: 1
DESCRIPTION: Dry run complete
COMPUTED VALUES:
  replicaCount: 3
  image:
    repository: 3llimi/devops-info-service
    tag: latest
  service:
    type: NodePort
    port: 80
    targetPort: 8000
    nodePort: 30080
```

---

## Bonus — Library Chart

### Why a Library Chart

Both `devops-python` and `devops-go` charts need identical label templates, name generation, and selector logic. Without a library chart this means copy-pasting `_helpers.tpl` — any change to label structure requires updating both charts manually.

The `common-lib` library chart solves this with the DRY principle: one source of truth for all shared templates.

### Library Chart Structure

```
k8s/common-lib/
├── Chart.yaml          # type: library — cannot be installed directly
└── templates/
    └── _helpers.tpl    # Shared: common.name, common.fullname,
                        #         common.chart, common.labels,
                        #         common.selectorLabels
```

**`Chart.yaml`:**
```yaml
apiVersion: v2
name: common-lib
description: Shared template library for DevOps course applications
type: library
version: 0.1.0
```

**Key difference from application charts:** `type: library` — Helm refuses to install it directly. It can only be used as a dependency.

### Shared Templates

| Template | Purpose |
|----------|---------|
| `common.name` | Chart name with optional override, truncated to 63 chars |
| `common.fullname` | `release-chart` format with optional full override |
| `common.chart` | `chart-version` string for `helm.sh/chart` label |
| `common.labels` | Full set of recommended Kubernetes labels |
| `common.selectorLabels` | Minimal labels for pod selection |

### Dependency Configuration

Both app charts declare `common-lib` as a dependency:

```yaml
# devops-python/Chart.yaml and devops-go/Chart.yaml
dependencies:
  - name: common-lib
    version: 0.1.0
    repository: "file://../common-lib"
```

`file://` prefix tells Helm to resolve the dependency from the local filesystem instead of a remote repository — correct for monorepo setups.

### Using Library Templates

Both charts reference shared templates identically:

```yaml
# deployment.yaml
metadata:
  name: {{ include "common.fullname" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
```

### Benefits

| Aspect | Without Library | With Library |
|--------|----------------|--------------|
| **Label consistency** | Manual sync between charts | Single definition |
| **Naming logic** | Duplicated in each chart | One implementation |
| **Maintenance** | Update N charts for 1 change | Update library once |
| **DRY** | ❌ Copy-paste | ✅ Shared templates |
| **Onboarding** | Learn each chart's helpers | Learn one library |

### Deployment Evidence

```bash
$ helm dependency update k8s/devops-python
Saving 1 charts
Deleting outdated charts

$ helm dependency update k8s/devops-go
Saving 1 charts
Deleting outdated charts

$ helm install devops-python k8s/devops-python
NAME: devops-python | STATUS: deployed | REVISION: 1

$ helm install devops-go k8s/devops-go
NAME: devops-go | STATUS: deployed | REVISION: 1

$ helm list
NAME          NAMESPACE  REVISION  STATUS    CHART               APP VERSION
devops-go     default    1         deployed  devops-go-0.1.0     1.0.0
devops-python default    1         deployed  devops-python-0.1.0 1.0.0
```

Both apps deployed successfully using shared templates from `common-lib`. The `devops-python` deployment additionally runs pre/post install hooks for lifecycle management.

---

## Summary

| Component | Details |
|-----------|---------|
| Helm version | v4.1.3 |
| Charts created | devops-python, devops-go, common-lib |
| Library chart | common-lib (shared labels, names, selectors) |
| Environments | dev (1 replica, minimal resources), prod (5 replicas, pinned tag) |
| Hooks | pre-install (weight -5), post-install (weight +5) |
| Hook policy | before-hook-creation |
| Releases deployed | 2 (devops-python, devops-go) |
| Total pods | 6 (3 per app) |
| Services | NodePort 30080 (Python), NodePort 30081 (Go) |