# Lab 11 — Kubernetes Secrets & HashiCorp Vault

## Task 1 — Kubernetes Secrets Fundamentals

### Creating a Secret

```bash
$ kubectl create secret generic app-credentials \
  --from-literal=username=admin \
  --from-literal=password=secret123

secret/app-credentials created
```

### Viewing the Secret

```bash
$ kubectl get secret app-credentials -o yaml

apiVersion: v1
data:
  password: c2VjcmV0MTIz
  username: YWRtaW4=
kind: Secret
metadata:
  creationTimestamp: "2026-03-16T02:00:32Z"
  name: app-credentials
  namespace: default
  resourceVersion: "3777"
  uid: 5d192aff-dc7a-4c04-b6ef-9864d300bc65
type: Opaque
```

### Decoding the Values

```powershell
# Decode username
[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String("YWRtaW4="))
admin

# Decode password
[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String("c2VjcmV0MTIz"))
secret123
```

### Base64 Encoding vs Encryption

**Base64 is encoding, NOT encryption.** The values are trivially reversible — anyone with `kubectl get secret` access can decode them instantly, as demonstrated above.

**What this means in practice:**
- Kubernetes Secrets are stored in etcd in base64-encoded form
- By default, etcd is NOT encrypted at rest
- Any user with RBAC access to `get secrets` can read all values
- Secrets are only obfuscated, not protected

**How to actually secure Kubernetes Secrets:**

1. **etcd encryption at rest** — Enable `EncryptionConfiguration` in the API server to encrypt secret data before writing to etcd. Requires control plane access (not available in managed clusters without extra config).

2. **RBAC restrictions** — Limit which service accounts and users can `get`/`list` secrets. Use least-privilege roles.

3. **External secret managers** — Use HashiCorp Vault, AWS Secrets Manager, or Azure Key Vault to store the actual values outside Kubernetes entirely. Kubernetes only gets a reference, not the value.

4. **Sealed Secrets** — Encrypt secrets client-side before committing to Git, only decryptable by the cluster controller.

**For production:** Always use an external secret manager like Vault (Task 3). Native Kubernetes Secrets are acceptable only for non-sensitive config or when combined with etcd encryption and strict RBAC.

---

## Task 2 — Helm-Managed Secrets

### Secret Template

**`k8s/devops-python/templates/secrets.yaml`:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "common.fullname" . }}-secret
  labels:
    {{- include "common.labels" . | nindent 4 }}
type: Opaque
stringData:
  username: {{ .Values.secret.username }}
  password: {{ .Values.secret.password }}
```

**Why `stringData` instead of `data`:** `stringData` accepts plain text and Kubernetes automatically base64-encodes it. Using `data` would require pre-encoding values in `values.yaml`, making them harder to read and maintain.

### Values Configuration

Added to `k8s/devops-python/values.yaml`:

```yaml
secret:
  username: "app-user"
  password: "changeme"
```

**Security note:** These are placeholder values. Real credentials are never committed to Git — they are injected at deploy time via `--set` flags or external secret management (Vault).

### Secret Injection in Deployment

Updated `k8s/devops-python/templates/deployment.yaml` to consume the secret via `envFrom`:

```yaml
        envFrom:
          - secretRef:
              name: {{ include "common.fullname" . }}-secret
```

This injects all keys from the secret as environment variables automatically. No need to list each key individually.

### Verification

```bash
$ kubectl exec -it devops-python-devops-python-7d9f7c46fb-2q9hg -c devops-python -- env | Select-String -Pattern "username|password"

password=changeme
username=app-user
```

Environment variables injected successfully. The secret values are available inside the container without being visible in `kubectl describe pod` output.

### Resource Limits

Resource requests and limits were already configured from Lab 10. Summary:

| Service | CPU Request | CPU Limit | RAM Request | RAM Limit |
|---------|------------|-----------|-------------|-----------|
| Python | 100m | 200m | 128Mi | 256Mi |
| Go | 50m | 100m | 64Mi | 128Mi |

**Requests vs Limits:**
- **Requests** — Minimum guaranteed resources. The scheduler uses this to find a node with enough capacity. Pod is placed only on nodes that can satisfy the request.
- **Limits** — Hard ceiling. If the container exceeds its memory limit, it is OOMKilled. If it exceeds CPU limit, it is throttled (not killed).

**How to choose values:** Start with requests at ~50% of observed average usage, limits at ~2x requests. Adjust based on monitoring data (Prometheus from Lab 8). Avoid setting limits too tight — it causes unnecessary throttling and restarts.

---

## Task 3 — HashiCorp Vault Integration

### Installation

HashiCorp's Helm repository is blocked in Russia (403 Forbidden). Installed via direct GitHub release download:

```bash
# Download Vault Helm chart from GitHub releases
Invoke-WebRequest -Uri "https://github.com/hashicorp/vault-helm/archive/refs/tags/v0.28.1.tar.gz" -OutFile "vault-helm.tar.gz"
tar -xzf vault-helm.tar.gz

# Install in dev mode with agent injector enabled
helm install vault ./vault-helm-0.28.1 \
  --set "server.dev.enabled=true" \
  --set "injector.enabled=true"
```

**Dev mode** auto-initializes and unseals Vault with a known root token. Never use dev mode in production — it stores data in memory and loses all secrets on restart.

### Vault Pods Running

```bash
$ kubectl get pods

NAME                                    READY   STATUS    AGE
vault-0                                 1/1     Running   63s
vault-agent-injector-5d48bf476c-fvnnm   1/1     Running   63s
```

Two components deployed:
- **`vault-0`** — The Vault server storing and serving secrets
- **`vault-agent-injector`** — Webhook that intercepts pod creation and injects the sidecar agent

### Vault Configuration

Exec into Vault pod and configure:

```bash
kubectl exec -it vault-0 -- /bin/sh
```

**1. KV secrets engine** — Already enabled at `secret/` path in dev mode:

```bash
# Create application secret
$ vault kv put secret/devops-python/config username="app-user" password="supersecret123"

========== Secret Path ==========
secret/data/devops-python/config
======= Metadata =======
Key              Value
created_time     2026-03-16T02:37:35.710459266Z
version          1

$ vault kv get secret/devops-python/config
====== Data ======
Key         Value
---         -----
password    supersecret123
username    app-user
```

**2. Kubernetes auth method:**

```bash
$ vault auth enable kubernetes
Success! Enabled kubernetes auth method at: kubernetes/

$ vault write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"
Success! Data written to: auth/kubernetes/config
```

**3. Policy — grants read access to the secret path:**

```bash
$ vault policy write devops-python - <<EOF
path "secret/data/devops-python/config" {
  capabilities = ["read"]
}
EOF
Success! Uploaded policy: devops-python
```

**4. Role — binds policy to Kubernetes service account:**

```bash
$ vault write auth/kubernetes/role/devops-python \
  bound_service_account_names=default \
  bound_service_account_namespaces=default \
  policies=devops-python \
  ttl=24h
Success! Data written to: auth/kubernetes/role/devops-python
```

### Vault Agent Sidecar Injection

Added annotations to the deployment pod template:

```yaml
annotations:
  vault.hashicorp.com/agent-inject: "true"
  vault.hashicorp.com/role: "devops-python"
  vault.hashicorp.com/agent-inject-secret-config: "secret/data/devops-python/config"
```

**How sidecar injection works:**
1. Pod is submitted to Kubernetes API
2. Vault Agent Injector webhook intercepts the request
3. Webhook mutates the pod spec — adds init container and sidecar container
4. Init container authenticates with Vault using the pod's service account token
5. Init container fetches secrets and writes them to a shared volume at `/vault/secrets/`
6. App container starts after init completes — secrets are already available
7. Sidecar container runs alongside app, renewing secrets before TTL expires

### Injection Evidence

Pods now show `2/2` — app container + vault-agent sidecar:

```bash
$ kubectl get pods
NAME                                          READY   STATUS    AGE
devops-python-devops-python-6448d5675-dmmqj   2/2     Running   61s
devops-python-devops-python-6448d5675-ngrr2   2/2     Running   44s
devops-python-devops-python-6448d5675-xbhxm   2/2     Running   53s
vault-0                                       1/1     Running   5m52s
vault-agent-injector-5d48bf476c-fvnnm         1/1     Running   5m52s
```

Secret file available at `/vault/secrets/config`:

```bash
$ kubectl exec -it devops-python-devops-python-6448d5675-dmmqj -c devops-python -- cat /vault/secrets/config

data: map[password:supersecret123 username:app-user]
metadata: map[created_time:2026-03-16T02:37:35.710459266Z ...]
```

---

## Bonus — Vault Agent Templates & Named Templates

### Custom Template Rendering

Added `agent-inject-template-config` annotation to render secrets in `.env` format instead of the default Vault response format:

```yaml
vault.hashicorp.com/agent-inject-template-config: |
  {{- with secret "secret/data/devops-python/config" -}}
  USERNAME={{ .Data.data.username }}
  PASSWORD={{ .Data.data.password }}
  APP_ENV=production
  {{- end -}}
```

**Note:** Vault Agent uses Go templating internally. In the Helm chart, the `{{ }}` syntax must be escaped using `{{ "{{" }}` and `{{ "}}" }}` to prevent Helm from interpreting them as Helm template directives.

**Rendered output in pod:**

```bash
$ kubectl exec -it devops-python-devops-python-5c8d4f7446-2n2n5 -c devops-python -- cat /vault/secrets/config

USERNAME=app-user
PASSWORD=supersecret123
APP_ENV=production
```

Clean `.env` format — directly sourceable by shell scripts or readable by apps expecting key=value files.

### Secret Rotation

Vault Agent automatically handles secret renewal:
- The sidecar container runs continuously alongside the app
- Before the lease TTL expires (24h in our config), the agent re-authenticates and fetches fresh values
- Updated values are written to `/vault/secrets/` without restarting the pod
- The `vault.hashicorp.com/agent-inject-command` annotation can trigger a process signal (e.g., `SIGHUP`) to notify the app of updated secrets

### Named Template for Environment Variables

Added `devops-python.envVars` named template to `_helpers.tpl`:

```yaml
{{/*
Common environment variables
*/}}
{{- define "devops-python.envVars" -}}
env:
  - name: APP_ENV
    value: {{ .Values.appEnv | default "production" }}
  - name: LOG_LEVEL
    value: {{ .Values.logLevel | default "INFO" }}
{{- end -}}
```

Referenced in `deployment.yaml`:

```yaml
        {{- include "devops-python.envVars" . | nindent 8 }}
```

**Verification:**

```bash
$ kubectl exec -it devops-python-devops-python-5c8d4f7446-2n2n5 -c devops-python -- env | Select-String -Pattern "APP_ENV|LOG_LEVEL"

LOG_LEVEL=INFO
APP_ENV=production
```

**Benefits of named templates for env vars:**
- DRY — define once, use in multiple deployments (e.g., canary, blue/green)
- Consistent — all deployments get the same base env vars
- Maintainable — change one template instead of every deployment manifest

---

## Security Analysis

### Kubernetes Secrets vs Vault

| Aspect | Kubernetes Secrets | HashiCorp Vault |
|--------|-------------------|-----------------|
| **Storage** | etcd (base64, not encrypted by default) | Encrypted storage backend |
| **Access control** | Kubernetes RBAC | Fine-grained Vault policies |
| **Audit logging** | Basic k8s audit log | Detailed per-secret access log |
| **Secret rotation** | Manual (redeploy) | Automatic with lease renewal |
| **Dynamic secrets** | Not supported | Generates credentials on demand |
| **Multi-cluster** | Per-cluster only | Centralized across clusters |
| **Complexity** | Low | High |
| **Setup time** | Seconds | Hours |

### When to Use Each

**Use Kubernetes Secrets when:**
- Small teams with simple deployments
- Non-sensitive config (feature flags, endpoints)
- etcd encryption is enabled
- Combined with strict RBAC
- Budget or complexity constraints prevent Vault

**Use Vault when:**
- Multiple teams or clusters sharing secrets
- Compliance requirements (SOC2, PCI-DSS, HIPAA)
- Need for audit trails per secret access
- Dynamic credentials (database passwords that rotate)
- Secrets need to be shared across non-Kubernetes workloads

### Production Recommendations

1. **Never commit real secrets to Git** — Use placeholder values in `values.yaml`, inject real values at deploy time via `--set` or Vault
2. **Enable etcd encryption** if using native Kubernetes Secrets in production
3. **Use Vault for sensitive data** — database passwords, API keys, certificates
4. **Restrict RBAC** — pods should only access their own secrets, not cluster-wide
5. **Set short TTLs** — 24h maximum for Vault leases, forces regular rotation
6. **Audit regularly** — Review Vault access logs to detect unauthorized access
7. **Use `hook-delete-policy: hook-succeeded`** on Helm hooks that handle secrets — don't leave job pods with secret env vars running

---

## Full Cluster State

```bash
$ kubectl get pods
NAME                                             READY   STATUS      RESTARTS   AGE
devops-go-devops-go-5f67859b64-9jqv6             1/1     Running     0          75m
devops-go-devops-go-5f67859b64-gm9cc             1/1     Running     0          75m
devops-go-devops-go-5f67859b64-nhw29             1/1     Running     0          75m
devops-python-devops-python-5c8d4f7446-2n2n5     2/2     Running     0          5m17s
devops-python-devops-python-5c8d4f7446-q5dcq     2/2     Running     0          5m
devops-python-devops-python-5c8d4f7446-vl485     2/2     Running     0          5m8s
devops-python-devops-python-post-install-8jtlr   0/1     Completed   0          75m
devops-python-devops-python-pre-install-6m9l8    0/1     Completed   0          75m
vault-0                                          1/1     Running     0          13m
vault-agent-injector-5d48bf476c-fvnnm            1/1     Running     0          13m
```

Python pods show `2/2` — app container + vault-agent sidecar running in each pod.

---

## Summary

| Component | Details |
|-----------|---------|
| Kubernetes Secret | `app-credentials` with username/password, base64 encoded |
| Helm Secret Template | `templates/secrets.yaml` using `stringData`, injected via `envFrom` |
| Vault version | v1.18 (dev mode) |
| Vault chart | hashicorp/vault v0.28.1 |
| KV engine | kv-v2 at `secret/devops-python/config` |
| Auth method | Kubernetes, bound to `default` service account |
| Policy | `devops-python` — read-only on config path |
| Injection method | Vault Agent sidecar, secrets at `/vault/secrets/config` |
| Template format | Custom `.env` format via `agent-inject-template-config` |
| Named template | `devops-python.envVars` for APP_ENV and LOG_LEVEL |
| Pod containers | 2/2 (app + vault-agent sidecar) |