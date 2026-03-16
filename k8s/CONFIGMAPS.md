# Lab 12 — ConfigMaps & Persistent Volumes

## Task 1 — Application Persistence Upgrade

### Visits Counter Implementation

The Python app was updated with a file-based visit counter. Two new components were added:

**Counter functions in `app_python/app.py`:**

```python
VISITS_FILE = os.getenv("VISITS_FILE", "/data/visits")
_visits_lock = threading.Lock()

def get_visits() -> int:
    try:
        with open(VISITS_FILE, "r") as f:
            return int(f.read().strip())
    except (FileNotFoundError, ValueError):
        return 0

def increment_visits() -> int:
    with _visits_lock:
        count = get_visits() + 1
        os.makedirs(os.path.dirname(VISITS_FILE), exist_ok=True)
        with open(VISITS_FILE, "w") as f:
            f.write(str(count))
        return count
```

**Thread safety:** `threading.Lock()` prevents race conditions when multiple concurrent requests try to read/increment/write simultaneously. Without the lock, two requests could both read `5`, both write `6`, losing one increment.

**Graceful fallback:** `get_visits()` catches `FileNotFoundError` (first run, no file yet) and `ValueError` (corrupted file) — returns 0 in both cases instead of crashing.

**`VISITS_FILE` env var:** Path is configurable via environment variable, defaulting to `/data/visits`. This allows different paths in Docker Compose vs Kubernetes without code changes.

### New Endpoints

**`GET /`** — Now includes `visits` field and updated endpoints list:
```json
{
  "visits": 3,
  "endpoints": [
    {"path": "/", "method": "GET", "description": "Service information"},
    {"path": "/health", "method": "GET", "description": "Health check"},
    {"path": "/visits", "method": "GET", "description": "Visit counter"}
  ]
}
```

**`GET /visits`** — Dedicated counter endpoint:
```json
{"visits": 3, "timestamp": "2026-03-16T03:45:33.899229+00:00"}
```

### Docker Compose Volume

Updated `monitoring/docker-compose.yml` to mount a bind volume for the visits file:

```yaml
volumes:
  loki-data:
  grafana-data:
  prometheus-data:

services:
  app-python:
    volumes:
      - ./data:/data
    environment:
      - VISITS_FILE=/data/visits
```

**Why bind mount instead of named volume:** Docker named volumes are owned by root when first created, causing `PermissionError` since the container runs as non-root `appuser`. A bind mount uses the host directory permissions which are writable by the container process.

### Local Testing Evidence

```bash
$ curl.exe http://localhost:8000/
{"visits":1,...}

$ curl.exe http://localhost:8000/
{"visits":2,...}

$ curl.exe http://localhost:8000/
{"visits":3,...}

$ curl.exe http://localhost:8000/visits
{"visits":3,"timestamp":"2026-03-16T03:45:33.899229+00:00"}
```

**Persistence across restart:**
```bash
$ docker compose restart app-python

$ curl.exe http://localhost:8000/visits
{"visits":3,"timestamp":"2026-03-16T03:46:59.126562+00:00"}

$ curl.exe http://localhost:8000/
{"visits":4,...}
```

Counter continued from 3 → 4 after restart, confirming persistence via bind mount volume.

---

## Task 2 — ConfigMaps

### ConfigMap Template Structure

**`k8s/devops-python/templates/configmap.yaml`** defines two ConfigMaps:

**1. File ConfigMap** — loads `config.json` from the chart's `files/` directory:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-config
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
  config.json: |-
{{ .Files.Get "files/config.json" | indent 4 }}
```

**2. Env ConfigMap** — key-value pairs for environment variables:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-env
data:
  APP_ENV: {{ .Values.appEnv | quote }}
  LOG_LEVEL: {{ .Values.logLevel | quote }}
  APP_NAME: "devops-info-service"
  APP_VERSION: "1.0.0"
```

### config.json Content

**`k8s/devops-python/files/config.json`:**

```json
{
  "app_name": "devops-info-service",
  "environment": "production",
  "version": "1.0.0",
  "features": {
    "visits_counter": true,
    "metrics_enabled": true,
    "json_logging": true
  },
  "settings": {
    "log_level": "INFO",
    "max_visits_file_size": "1MB"
  }
}
```

### Mounting ConfigMap as File

In `deployment.yaml` — volume mount and volume definition:

```yaml
        volumeMounts:
        - name: config-volume
          mountPath: /config
        - name: data-volume
          mountPath: /data
      volumes:
      - name: config-volume
        configMap:
          name: {{ include "common.fullname" . }}-config
      - name: data-volume
        persistentVolumeClaim:
          claimName: {{ include "common.fullname" . }}-data
```

The entire ConfigMap is mounted as a directory at `/config`. The `config.json` key becomes the file `/config/config.json`.

### ConfigMap as Environment Variables

The env ConfigMap is injected via `envFrom`:

```yaml
        envFrom:
          - secretRef:
              name: {{ include "common.fullname" . }}-secret
          - configMapRef:
              name: {{ include "common.fullname" . }}-env
```

All keys from the ConfigMap become environment variables automatically.

### Verification

**ConfigMap resources:**
```bash
$ kubectl get configmap,pvc
NAME                                           DATA   AGE
configmap/devops-python-devops-python-config   1      2m43s
configmap/devops-python-devops-python-env      4      32m
configmap/kube-root-ca.crt                     1      3h12m
```

**File mounted inside pod:**
```bash
$ kubectl exec -it devops-python-devops-python-7497cd898d-dmpzx -c devops-python -- cat /config/config.json
{
  "app_name": "devops-info-service",
  "environment": "production",
  "version": "1.0.0",
  "features": {
    "visits_counter": true,
    "metrics_enabled": true,
    "json_logging": true
  },
  "settings": {
    "log_level": "INFO",
    "max_visits_file_size": "1MB"
  }
}
```

**Environment variables injected:**
```bash
$ kubectl exec -it devops-python-devops-python-7497cd898d-dmpzx -c devops-python -- env | Select-String -Pattern "APP_ENV|LOG_LEVEL|APP_NAME|APP_VERSION"

APP_ENV=production
APP_NAME=devops-info-service
APP_VERSION=1.0.0
LOG_LEVEL=INFO
```

---

## Task 3 — Persistent Volumes

### PVC Template

**`k8s/devops-python/templates/pvc.yaml`:**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "common.fullname" . }}-data
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.persistence.size }}
  {{- if .Values.persistence.storageClass }}
  storageClassName: {{ .Values.persistence.storageClass }}
  {{- end }}
```

**Values:**
```yaml
persistence:
  enabled: true
  size: 100Mi
  storageClass: ""
```

### Access Modes

**`ReadWriteOnce` (RWO):** The volume can be mounted read-write by a single node. Suitable for our visits counter since all pods run on the same minikube node.

Other access modes for reference:
- `ReadWriteMany` (RWX) — multiple nodes can mount read-write (NFS, cloud file storage)
- `ReadOnlyMany` (ROX) — multiple nodes can mount read-only

**Storage class:** Empty string uses the cluster default (`standard` in minikube, which provisions hostPath volumes automatically). In production, you would specify a cloud storage class like `gp3` (AWS) or `premium-rrs` (GCP).

### PVC Status

```bash
$ kubectl get pvc
NAME                                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
devops-python-devops-python-data       Bound    pvc-41b2a312-48e8-4b7c-8f3a-7d1cc7d0986a   100Mi      RWO            standard       32m
```

`Bound` status confirms minikube's default storage provisioner automatically created a PersistentVolume and bound it to the claim.

### Persistence Test

**Visits before pod deletion:**
```bash
$ curl.exe http://127.0.0.1:51025/
{"visits":2,...}
$ curl.exe http://127.0.0.1:51025/
{"visits":3,...}
$ curl.exe http://127.0.0.1:51025/
{"visits":4,...}
$ curl.exe http://127.0.0.1:51025/visits
{"visits":4,"timestamp":"2026-03-16T04:09:58.398057+00:00"}
```

**Pod deletion:**
```bash
$ kubectl delete pod devops-python-devops-python-7497cd898d-7d5pb
pod "devops-python-devops-python-7497cd898d-7d5pb" deleted
```

**New pod started:**
```bash
$ kubectl get pods
NAME                                           READY   STATUS    AGE
devops-python-devops-python-7497cd898d-dmpzx   2/2     Running   5s
devops-python-devops-python-7497cd898d-szpxl   2/2     Running   10m
devops-python-devops-python-7497cd898d-xpdw4   2/2     Running   10m
```

**Visits after pod deletion — data survived:**
```bash
$ curl.exe http://127.0.0.1:51025/visits
{"visits":4,"timestamp":"2026-03-16T04:10:12.974397+00:00"}
```

Counter preserved at 4 after the pod was deleted and replaced. The PVC outlives individual pods — data persists as long as the PVC exists.

---

## Bonus — ConfigMap Hot Reload

### Default Update Behavior

ConfigMap mounted as a directory volume updates automatically without pod restart. Tested by editing the ConfigMap directly:

```bash
kubectl edit configmap devops-python-devops-python-config
# Changed "environment": "production" → "environment": "staging"
```

After ~60 seconds (kubelet sync period):

```bash
$ kubectl exec -it devops-python-devops-python-7497cd898d-dmpzx -c devops-python -- cat /config/config.json
{
  "app_name": "devops-info-service",
  "environment": "staging",
  ...
}
```

File updated inside the pod without any restart. The kubelet polls for ConfigMap changes every 60 seconds by default (configurable via `--sync-frequency`).

### subPath Limitation

When mounting a ConfigMap using `subPath`, the file is copied once at pod creation and **never updated**, even when the ConfigMap changes. This is because `subPath` mounts create a direct bind mount to the file, bypassing the symlink mechanism that enables auto-updates.

**When to use subPath:** When you need to inject a single file into a directory that contains other files (to avoid replacing the entire directory). Accept the trade-off that the file won't auto-update.

**When to avoid subPath:** When you need hot reload capability. Use full directory mounts instead.

### Checksum Annotation Pattern

The deployment includes a checksum annotation that triggers pod restarts when the ConfigMap content changes via Helm:

```yaml
annotations:
  checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
```

**How it works:** Every time `helm upgrade` runs, Helm renders the ConfigMap template and computes its SHA256 hash. If the hash changes (because `files/config.json` was modified), the annotation value changes, which triggers a rolling restart of all pods.

**Why this matters:** Without this annotation, `helm upgrade` would update the ConfigMap but pods would keep running with stale in-memory config until the kubelet sync catches up. The checksum annotation ensures pods always restart with fresh config after a Helm upgrade.

**Tested behavior:** When `kubectl edit` modified the ConfigMap outside Helm, the conflict was resolved by deleting the ConfigMap and running `helm upgrade` — Helm recreated it with `production` values and the checksum annotation ensured consistency going forward.

### Reload Approach Comparison

| Approach | Complexity | Restart Required | Delay |
|----------|-----------|-----------------|-------|
| Directory mount (default) | Low | No | 60s kubelet sync |
| Checksum annotation | Low | Yes (rolling) | Immediate on `helm upgrade` |
| subPath mount | Low | Yes (manual) | Never auto-updates |
| Stakater Reloader | Medium | Yes (automatic) | Seconds after CM change |
| App file watching | High | No | Milliseconds |

For this lab, the checksum annotation approach was implemented — it's the industry standard Helm pattern that balances simplicity with correctness.

---

## ConfigMap vs Secret

| Aspect | ConfigMap | Secret |
|--------|-----------|--------|
| **Purpose** | Non-sensitive configuration | Sensitive credentials |
| **Storage** | Plain text in etcd | Base64-encoded in etcd |
| **Use cases** | App config, feature flags, env settings | Passwords, API keys, TLS certs |
| **Git safe** | ✅ Yes (no sensitive data) | ❌ No (encode only, not encrypt) |
| **Vault integration** | Not needed | Recommended for production |
| **Size limit** | 1MB | 1MB |
| **Access control** | Standard RBAC | Standard RBAC (same as ConfigMap) |

**Use ConfigMap when:** The data is non-sensitive and safe to store in version control — application settings, feature flags, log levels, connection strings without credentials, environment names.

**Use Secret when:** The data must not be exposed — passwords, tokens, API keys, TLS private keys, database credentials. For production, combine Secrets with Vault (Lab 11) to avoid storing sensitive data in etcd at all.

**Key insight:** Kubernetes Secrets are NOT more secure than ConfigMaps by default — both are stored in etcd with the same access controls. The distinction is semantic and tooling-based. Real security comes from etcd encryption at rest, RBAC policies, and external secret managers like Vault.

---

## Full Cluster State

```bash
$ kubectl get configmap,pvc
NAME                                           DATA   AGE
configmap/devops-python-devops-python-config   1      2m43s
configmap/devops-python-devops-python-env      4      32m
configmap/kube-root-ca.crt                     1      3h12m

NAME                                                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/devops-python-devops-python-data   Bound    pvc-41b2a312-48e8-4b7c-8f3a-7d1cc7d0986a   100Mi      RWO            standard       32m
```

---

## Summary

| Component | Details |
|-----------|---------|
| Visits counter | File-based, thread-safe, configurable path via env var |
| New endpoint | `GET /visits` returns current count and timestamp |
| Docker volume | Bind mount `./data:/data` for local persistence |
| ConfigMap (file) | `config.json` mounted at `/config/config.json` |
| ConfigMap (env) | 4 keys injected as env vars via `envFrom.configMapRef` |
| PVC | 100Mi, ReadWriteOnce, standard storage class, Bound |
| PVC mount | `/data` directory — visits file survives pod deletion |
| Hot reload | Directory mount auto-updates in ~60s (kubelet sync) |
| Checksum annotation | Triggers rolling restart when ConfigMap changes via Helm |
| subPath limitation | Does not auto-update — documented and avoided |