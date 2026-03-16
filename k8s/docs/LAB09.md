# Lab 9 — Kubernetes Fundamentals

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        minikube cluster                             │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   ingress-nginx controller                   │   │
│  │                                                             │   │
│  │   https://devops.local/app1 ──► devops-python-service:80   │   │
│  │   https://devops.local/app2 ──► devops-go-service:80       │   │
│  │                    TLS: devops-tls secret                   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌──────────────────────────┐  ┌──────────────────────────┐        │
│  │  devops-python           │  │  devops-go               │        │
│  │  Deployment              │  │  Deployment              │        │
│  │  3 replicas              │  │  3 replicas              │        │
│  │  image: devops-info-     │  │  image: devops-go-       │        │
│  │  service:latest          │  │  service:latest          │        │
│  │  port: 8000              │  │  port: 8080              │        │
│  │  CPU: 100m-200m          │  │  CPU: 50m-100m           │        │
│  │  MEM: 128Mi-256Mi        │  │  MEM: 64Mi-128Mi         │        │
│  └──────────────────────────┘  └──────────────────────────┘        │
│                                                                     │
│  ┌──────────────────────────┐  ┌──────────────────────────┐        │
│  │  devops-python-service   │  │  devops-go-service       │        │
│  │  ClusterIP :80           │  │  ClusterIP :80           │        │
│  └──────────────────────────┘  └──────────────────────────┘        │
└─────────────────────────────────────────────────────────────────────┘
```

**How it works:**
- Ingress controller handles all external HTTPS traffic with TLS termination
- Path-based routing forwards `/app1` to the Python service and `/app2` to the Go service
- Each service uses ClusterIP and load-balances across 3 pod replicas
- Liveness and readiness probes on `/health` ensure only healthy pods receive traffic

---

## Task 1 — Local Kubernetes Setup

### Tool Choice: minikube

**Why minikube over kind:**
- Full-featured local Kubernetes with addons (Ingress, metrics-server, dashboard)
- Simpler addon management (`minikube addons enable ingress`)
- Better documentation for beginners
- Docker driver works seamlessly with existing Docker Desktop on Windows

**Why Docker driver over VirtualBox:**
- Hyper-V is active on this machine (required for Docker Desktop and WSL2)
- VirtualBox cannot boot 64-bit VMs when Hyper-V is active
- Docker driver runs the minikube node as a Docker container — no conflict

### Installation

**kubectl** was already installed. **minikube** was installed via winget:

```
winget install Kubernetes.minikube
# Installed to: C:\Program Files\Kubernetes\Minikube\minikube.exe
```

```
minikube version: v1.38.1
commit: c93a4cb9311efc66b90d33ea03f75f2c4120e9b0
```

### Cluster Setup

```bash
minikube start --driver=docker
```

```
😄  minikube v1.38.1 on Microsoft Windows 11 Pro 25H2
✨  Using the docker driver based on user configuration
📌  Using Docker Desktop driver with root privileges
👍  Starting "minikube" primary control-plane node in "minikube" cluster
🚜  Pulling base image v0.0.50 ...
🔥  Creating docker container (CPUs=2, Memory=8100MB) ...
🐳  Preparing Kubernetes v1.35.1 on Docker 29.2.1 ...
🔗  Configuring bridge CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```

### Cluster Verification

```bash
$ kubectl cluster-info

Kubernetes control plane is running at https://127.0.0.1:11819
CoreDNS is running at https://127.0.0.1:11819/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

```bash
$ kubectl get nodes

NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane   26s   v1.35.1
```

```bash
$ kubectl get namespaces

NAME              STATUS   AGE
default           Active   30s
kube-node-lease   Active   30s
kube-public       Active   30s
kube-system       Active   30s
```

**Single-node cluster** running Kubernetes v1.35.1. The control plane and worker roles are combined in one node for local development — this is normal for minikube.

---

## Task 2 — Application Deployment

### Manifest: `k8s/deployment.yml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-python
  labels:
    app: devops-python
spec:
  replicas: 3
  selector:
    matchLabels:
      app: devops-python
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: devops-python
    spec:
      containers:
      - name: devops-python
        image: 3llimi/devops-info-service:latest
        ports:
        - containerPort: 8000
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
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

### Key Configuration Decisions

**Replicas: 3** — Minimum for high availability. Allows one pod to be down during updates (maxUnavailable: 0) without losing capacity.

**RollingUpdate with maxUnavailable: 0** — Zero downtime updates. Kubernetes always keeps all 3 pods running, adding the new pod first before removing an old one.

**Resource requests vs limits:**
- Requests (100m CPU, 128Mi RAM) — what the scheduler uses to place the pod on a node
- Limits (200m CPU, 256Mi RAM) — hard ceiling to prevent resource starvation
- 2x ratio between request and limit allows burst headroom

**Liveness probe** — Restarts the container if `/health` fails 3 consecutive times. initialDelaySeconds: 10 gives the app time to start before probing begins.

**Readiness probe** — Removes the pod from the service load balancer if `/health` fails. initialDelaySeconds: 5 is shorter than liveness because we want to detect readiness faster.

**Non-root user** — Already baked into the Docker image from Lab 2 (`appuser`).

### Deployment Evidence

```bash
$ kubectl apply -f k8s/deployment.yml
deployment.apps/devops-python created

$ kubectl get deployments
NAME            READY   UP-TO-DATE   AVAILABLE   AGE
devops-python   3/3     3            3           29s

$ kubectl get pods
NAME                             READY   STATUS    RESTARTS   AGE
devops-python-5f79479cfd-tz5wm   1/1     Running   0          53s
devops-python-5f79479cfd-vhkc8   1/1     Running   0          53s
devops-python-5f79479cfd-vpm4s   1/1     Running   0          53s
```

```bash
$ kubectl describe deployment devops-python

Name:                   devops-python
Namespace:              default
CreationTimestamp:      Sun, 15 Mar 2026 06:15:31 +0300
Labels:                 app=devops-python
Replicas:               3 desired | 3 updated | 3 total | 3 available | 0 unavailable
StrategyType:           RollingUpdate
RollingUpdateStrategy:  0 max unavailable, 1 max surge
Pod Template:
  Labels:  app=devops-python
  Containers:
   devops-python:
    Image:      3llimi/devops-info-service:latest
    Port:       8000/TCP
    Limits:
      cpu:     200m
      memory:  256Mi
    Requests:
      cpu:         100m
      memory:      128Mi
    Liveness:      http-get http://:8000/health delay=10s timeout=1s period=10s #success=1 #failure=3
    Readiness:     http-get http://:8000/health delay=5s timeout=1s period=5s #success=1 #failure=3
Conditions:
  Type           Status  Reason
  ----           ------  ------
  Available      True    MinimumReplicasAvailable
  Progressing    True    NewReplicaSetAvailable
Events:
  Normal  ScalingReplicaSet  52s  deployment-controller  Scaled up replica set devops-python-5f79479cfd from 0 to 3
```

---

## Task 3 — Service Configuration

### Manifest: `k8s/service.yml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-python-service
  labels:
    app: devops-python
spec:
  type: ClusterIP
  selector:
    app: devops-python
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8000
```

**Service type:** Started as NodePort for initial testing, then changed to ClusterIP once Ingress was added. Ingress handles all external traffic — ClusterIP is sufficient and more secure (not exposed directly on node ports).

**Label selector** `app: devops-python` — Must match the labels on the Deployment's pod template. This is how Kubernetes knows which pods belong to this service.

**Port mapping** — Service listens on port 80, forwards to container port 8000. Standard HTTP port externally, app-specific port internally.

### Service Evidence

```bash
$ kubectl apply -f k8s/service.yml
service/devops-python-service created

$ kubectl get services
NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
devops-python-service   NodePort    10.105.70.60   <none>        80:30080/TCP   0s
kubernetes              ClusterIP   10.96.0.1      <none>        443/TCP        2m43s

$ kubectl get endpoints
NAME                    ENDPOINTS                                         AGE
devops-python-service   10.244.0.3:8000,10.244.0.4:8000,10.244.0.5:8000   0s
kubernetes              192.168.49.2:8443                                 2m43s
```

All 3 pod IPs registered as endpoints — load balancing is active across all replicas.

### Connectivity Verification (NodePort)

Initial testing via `minikube service` tunnel (before switching to Ingress):

```bash
$ minikube service devops-python-service --url
http://127.0.0.1:40639

$ curl.exe http://127.0.0.1:40639/health
{"status":"healthy","timestamp":"2026-03-15T03:19:30.871577+00:00","uptime_seconds":218}

$ curl.exe http://127.0.0.1:40639/
{"service":{"name":"devops-info-service","version":"1.0.0","description":"DevOps course info
service","framework":"FastAPI"},"system":{"hostname":"devops-python-5f79479cfd-vpm4s",...}}
```

Both endpoints responding with HTTP 200. The hostname in the response (`devops-python-5f79479cfd-vpm4s`) confirms the request hit one of the 3 pods.

---

## Task 4 — Scaling and Updates

### Scaling to 5 Replicas

```bash
$ kubectl scale deployment/devops-python --replicas=5

$ kubectl get pods -w
NAME                             READY   STATUS    RESTARTS   AGE
devops-python-5f79479cfd-98m4p   1/1     Running   0          56s
devops-python-5f79479cfd-cv9w2   1/1     Running   0          56s
devops-python-5f79479cfd-tz5wm   1/1     Running   0          5m23s
devops-python-5f79479cfd-vhkc8   1/1     Running   0          5m23s
devops-python-5f79479cfd-vpm4s   1/1     Running   0          5m23s
```

All 5 pods running — 2 new pods started alongside the original 3.

### Rolling Update

Updated `k8s/deployment.yml` image tag from `latest` to the pinned CalVer tag from Lab 3:

```yaml
image: 3llimi/devops-info-service:2026.02.11-89e5033
```

```bash
$ kubectl apply -f k8s/deployment.yml
deployment.apps/devops-python configured

$ kubectl rollout status deployment/devops-python
Waiting for deployment "devops-python" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 1 old replicas are pending termination...
Waiting for deployment "devops-python" rollout to finish: 1 old replicas are pending termination...
deployment "devops-python" successfully rolled out
```

**Why zero downtime:** `maxUnavailable: 0` ensures Kubernetes only terminates an old pod after a new one passes its readiness probe. Traffic always has healthy pods to serve.

### Rollback

```bash
$ kubectl rollout history deployment/devops-python
deployment.apps/devops-python
REVISION  CHANGE-CAUSE
1         <none>
2         <none>

$ kubectl rollout undo deployment/devops-python
deployment.apps/devops-python rolled back

$ kubectl rollout status deployment/devops-python
Waiting for deployment "devops-python" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "devops-python" rollout to finish: 1 old replicas are pending termination...
deployment "devops-python" successfully rolled out

$ kubectl get pods
NAME                             READY   STATUS    RESTARTS   AGE
devops-python-5f79479cfd-fhz7n   1/1     Running   0          49s
devops-python-5f79479cfd-gzr27   1/1     Running   0          58s
devops-python-5f79479cfd-tb595   1/1     Running   0          39s
```

Rollback restored revision 1 (`latest` tag). Kubernetes keeps the previous ReplicaSet around specifically to enable instant rollbacks — no re-pull needed.

---

## Bonus — Ingress with TLS

### Architecture

```
External HTTPS Request
        │
        ▼
  minikube tunnel
        │
        ▼
ingress-nginx controller (port 443)
        │
        ├── /app1 ──► devops-python-service:80 ──► pods :8000
        └── /app2 ──► devops-go-service:80    ──► pods :8080

TLS termination at Ingress using self-signed cert for devops.local
```

### Second App Deployment

**`k8s/deployment-go.yml`:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-go
  labels:
    app: devops-go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: devops-go
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: devops-go
    spec:
      containers:
      - name: devops-go
        image: 3llimi/devops-go-service:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 3
```

**Go app uses lower resources** — compiled Go binary is significantly lighter than Python + uvicorn. Half the CPU and memory limits compared to the Python app.

**`k8s/service-go.yml`:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-go-service
  labels:
    app: devops-go
spec:
  type: ClusterIP
  selector:
    app: devops-go
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

### Ingress Controller Setup

```bash
$ minikube addons enable ingress

💡  ingress is an addon maintained by Kubernetes.
    ▪ Using image registry.k8s.io/ingress-nginx/controller:v1.14.3
🔎  Verifying ingress addon...
🌟  The 'ingress' addon is enabled

$ kubectl get pods -n ingress-nginx
NAME                                        READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-9rlmb        0/1     Completed   0          52s
ingress-nginx-admission-patch-cfwj7         0/1     Completed   1          52s
ingress-nginx-controller-596f8778bc-r9rlt   1/1     Running     0          52s
```

### TLS Certificate

```bash
# Generate self-signed certificate for devops.local
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout k8s/tls.key -out k8s/tls.crt \
  -subj "/CN=devops.local/O=devops.local"

# Create Kubernetes TLS secret
kubectl create secret tls devops-tls --key k8s/tls.key --cert k8s/tls.crt

$ kubectl get secret devops-tls
NAME         TYPE                DATA   AGE
devops-tls   kubernetes.io/tls   2      4s
```

The TLS secret stores the certificate and private key as base64-encoded data. Ingress references this secret by name for TLS termination.

### Ingress Manifest: `k8s/ingress.yml`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - devops.local
    secretName: devops-tls
  rules:
  - host: devops.local
    http:
      paths:
      - path: /app1
        pathType: Prefix
        backend:
          service:
            name: devops-python-service
            port:
              number: 80
      - path: /app2
        pathType: Prefix
        backend:
          service:
            name: devops-go-service
            port:
              number: 80
```

**`rewrite-target: /`** — Strips the path prefix before forwarding. A request to `/app1/health` reaches the backend as `/health`, which is what the app expects.

**`ingressClassName: nginx`** — Explicitly selects the nginx Ingress controller. Required in Kubernetes 1.18+ where multiple Ingress controllers can coexist.

### Ingress Evidence

```bash
$ kubectl apply -f k8s/ingress.yml
ingress.networking.k8s.io/devops-ingress created

$ kubectl describe ingress devops-ingress
Name:             devops-ingress
Namespace:        default
Ingress Class:    nginx
TLS:
  devops-tls terminates devops.local
Rules:
  Host          Path  Backends
  ----          ----  --------
  devops.local
                /app1   devops-python-service:80 (10.244.0.11:8000,10.244.0.12:8000,10.244.0.13:8000)
                /app2   devops-go-service:80 (10.244.0.17:8080,10.244.0.18:8080,10.244.0.19:8080)
Annotations:    nginx.ingress.kubernetes.io/rewrite-target: /
Events:
  Normal  Sync  2s  nginx-ingress-controller  Scheduled for sync
```

### Access Setup

```bash
# Start tunnel in Administrator terminal
minikube tunnel

# Add devops.local to hosts file
Add-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value "127.0.0.1 devops.local"

$ kubectl get ingress
NAME             CLASS   HOSTS          ADDRESS        PORTS     AGE
devops-ingress   nginx   devops.local   192.168.49.2   80, 443   69s
```

### HTTPS Routing Verification

```bash
$ curl.exe -k https://devops.local/app1
{"service":{"name":"devops-info-service","version":"1.0.0","description":"DevOps course info
service","framework":"FastAPI"},"system":{"hostname":"devops-python-5f79479cfd-gzr27",
"platform":"Linux","platform_version":"Linux-5.15.167.4-microsoft-standard-WSL2-x86_64-with-glibc2.41",
"architecture":"x86_64","cpu_count":12,"python_version":"3.13.12"},...}

$ curl.exe -k https://devops.local/app2
{"service":{"name":"devops-info-service","version":"1.0.0","description":"DevOps course info
service","framework":"Go net/http"},"system":{"hostname":"devops-go-74c9c74457-wtgkz",
"platform":"linux","platform_version":"linux-amd64","architecture":"amd64","cpu_count":12,
"go_version":"go1.25.6"},...}
```

- `/app1` → Python app confirmed (`framework: FastAPI`, hostname `devops-python-*`)
- `/app2` → Go app confirmed (`framework: Go net/http`, hostname `devops-go-*`)

Both routes working over HTTPS with TLS termination at the Ingress layer.

### Ingress vs NodePort

| Aspect | NodePort | Ingress |
|--------|----------|---------|
| **Protocol** | L4 (TCP) | L7 (HTTP/HTTPS) |
| **Routing** | One port per service | Path/host-based routing |
| **TLS** | Not supported | Native TLS termination |
| **Port range** | 30000-32767 | Standard 80/443 |
| **Multiple apps** | Multiple ports needed | Single entry point |
| **Cost** | Free | Requires Ingress controller |

Ingress is the production standard — one entry point, path-based routing, TLS in one place.

---

## Full Cluster State

```bash
$ kubectl get all

NAME                                 READY   STATUS    RESTARTS   AGE
pod/devops-go-74c9c74457-lw8jp       1/1     Running   0          33s
pod/devops-go-74c9c74457-n9tl2       1/1     Running   0          33s
pod/devops-go-74c9c74457-wtgkz       1/1     Running   0          33s
pod/devops-python-5f79479cfd-fhz7n   1/1     Running   0          4m36s
pod/devops-python-5f79479cfd-gzr27   1/1     Running   0          4m45s
pod/devops-python-5f79479cfd-tb595   1/1     Running   0          4m26s

NAME                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/devops-go-service       ClusterIP   10.105.65.96   <none>        80/TCP    1s
service/devops-python-service   ClusterIP   10.105.70.60   <none>        80/TCP    10m
service/kubernetes              ClusterIP   10.96.0.1      <none>        443/TCP   12m

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/devops-go       3/3     3            3           33s
deployment.apps/devops-python   3/3     3            3           8m36s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/devops-go-74c9c74457       3         3         3       33s
replicaset.apps/devops-python-5f79479cfd   3         3         3       8m36s
replicaset.apps/devops-python-786dd8bf99   0         0         0       2m23s
```

The old ReplicaSet (`devops-python-786dd8bf99`, 0/0/0) is the pinned tag from the rolling update — kept by Kubernetes to enable instant rollback.

---

## Production Considerations

### Health Checks

**Why both liveness and readiness probes:**
- **Liveness** catches a deadlocked app that is running but not responding — restarts the container
- **Readiness** catches a temporarily overloaded or starting app — removes it from load balancing without restarting

Using only liveness would cause unnecessary restarts during slow startup or temporary spikes. Using only readiness would leave broken pods in the cluster forever.

**Why `/health` endpoint:**
- Lightweight — no database calls, no external dependencies
- Returns quickly (under 5ms) — won't cause probe timeouts
- Already implemented from Lab 1

### Resource Limits Rationale

| Service | CPU Request | CPU Limit | RAM Request | RAM Limit |
|---------|------------|-----------|-------------|-----------|
| Python | 100m | 200m | 128Mi | 256Mi |
| Go | 50m | 100m | 64Mi | 128Mi |

**Go uses half the resources** — compiled binary with no interpreter overhead. Python needs uvicorn + FastAPI + Python runtime. These values were chosen based on observed usage during testing.

**Why set limits at all:** Without limits, a misbehaving pod can consume all node resources, starving other pods. Limits enforce the principle of least privilege at the resource level.

### Production Improvements

**What would be added in production:**

1. **Namespace isolation** — Separate namespace per environment (dev/staging/prod) to prevent accidental cross-environment access
2. **Horizontal Pod Autoscaler (HPA)** — Auto-scale based on CPU/memory metrics instead of manual `kubectl scale`
3. **PodDisruptionBudget** — Guarantee minimum availability during node maintenance
4. **Network Policies** — Restrict pod-to-pod communication (currently all pods can talk to all pods)
5. **Secrets management** — Use Vault (Lab 11) instead of plain Kubernetes secrets for sensitive data
6. **Image digest pinning** — Use `image@sha256:...` instead of tags to prevent supply chain attacks
7. **RBAC** — Role-based access control to restrict who can deploy or view resources
8. **Real TLS certificates** — cert-manager + Let's Encrypt instead of self-signed certificates

### Monitoring and Observability

The Prometheus + Loki + Grafana stack from Labs 7 and 8 integrates naturally with Kubernetes:
- **Prometheus** can scrape pod metrics via service discovery using Kubernetes SD configs
- **Loki + Promtail** can collect pod logs via the Docker socket or node log paths
- **Grafana** dashboards already configured from previous labs

In production, these would be deployed as Kubernetes workloads themselves, forming a full in-cluster observability stack.

---

## Challenges & Solutions

**Challenge 1: VirtualBox vs Hyper-V conflict**

minikube defaulted to the VirtualBox driver, which fails when Hyper-V is active (required for Docker Desktop and WSL2). Error: `VirtualBox won't boot a 64bits VM when Hyper-V is activated`.

Fixed by using `--driver=docker`, which runs the minikube node as a Docker container — no VM needed, no conflict with Hyper-V.

**Challenge 2: Leftover virtualbox cluster**

After the VirtualBox failure, minikube left a broken cluster state. `minikube start --driver=docker` failed with: `The existing "minikube" cluster was created using the "virtualbox" driver, which is incompatible with requested "docker" driver.`

Fixed by running `minikube delete` to clean up the broken state before starting fresh.

**Challenge 3: PowerShell `curl` is not real curl**

`curl -k https://devops.local/app1` failed because PowerShell aliases `curl` to `Invoke-WebRequest`, which uses different flags. `-k` and `-SkipCertificateCheck` are not available in older PowerShell versions.

Fixed by using `curl.exe` which calls the actual Windows curl binary and supports standard curl flags including `-k` for skipping TLS verification.

**Challenge 4: minikube tunnel required for Ingress**

After setting up Ingress, the address `192.168.49.2` was the internal minikube IP, not accessible from the Windows host. Ingress wasn't reachable without the tunnel.

Fixed by running `minikube tunnel` in an Administrator terminal, which creates a network route from `127.0.0.1` into the minikube cluster, making Ingress accessible at `https://devops.local`.

---

## Summary

| Component | Details |
|-----------|---------|
| Cluster | minikube v1.38.1, Kubernetes v1.35.1, Docker driver |
| Python app | 3 replicas, 100m-200m CPU, 128-256Mi RAM |
| Go app | 3 replicas, 50m-100m CPU, 64-128Mi RAM |
| Services | ClusterIP (both apps, routed via Ingress) |
| Ingress | nginx, path-based routing, TLS with self-signed cert |
| Health checks | Liveness + readiness probes on `/health` |
| Scaling | Demonstrated: 3 → 5 → 3 replicas |
| Updates | Rolling update with zero downtime confirmed |
| Rollback | `kubectl rollout undo` demonstrated |
| TLS | Self-signed cert for `devops.local`, 365 days |
| HTTPS routes | `/app1` → Python, `/app2` → Go |