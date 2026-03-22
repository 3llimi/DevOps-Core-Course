# Lab 15 — StatefulSet Documentation

## 1. StatefulSet Overview

### Why StatefulSet?

Deployments are designed for stateless applications where pods are interchangeable. StatefulSets are designed for applications that need one or more of the following guarantees:

**Stable, unique network identifiers** — Each pod gets a persistent DNS name based on its ordinal index (`pod-0`, `pod-1`, `pod-2`). This name is stable across restarts. A Deployment pod that restarts gets a new random name and IP; a StatefulSet pod that restarts gets the same name and DNS entry.

**Stable, persistent storage** — Each pod gets its own PersistentVolumeClaim via `volumeClaimTemplates`. The PVC is tied to the pod's identity, not the pod's lifetime — deleting a pod does not delete its PVC. When the pod is recreated, it reattaches to the same volume.

**Ordered, graceful deployment and scaling** — Pods are created in order (0 → 1 → 2) and each must be Running and Ready before the next starts. Scaling down happens in reverse order (2 → 1 → 0). This is essential for distributed systems like databases where each node must initialize before the next joins.

### Key Differences: Deployment vs StatefulSet

| Feature | Deployment | StatefulSet |
|---------|------------|-------------|
| Pod names | Random suffix (`app-7d9f7c46fb-2q9hg`) | Ordered index (`app-0`, `app-1`, `app-2`) |
| Storage | Shared PVC or no PVC | Per-pod PVC via `volumeClaimTemplates` |
| Scaling order | Any order, parallel | Ordered (0→1→2 up, 2→1→0 down) |
| Network identity | Random IP, random DNS | Stable DNS via headless service |
| Pod restart | New name, new IP | Same name, same PVC, new IP |
| Use case | Stateless apps (web servers, APIs) | Stateful apps (databases, queues) |

### When to Use Deployment vs StatefulSet

**Use Deployment when:**
- The application has no local state (all state in external DB or cache)
- Any pod can handle any request
- Pods are truly interchangeable
- Example: REST APIs, web frontends, microservices

**Use StatefulSet when:**
- Each instance needs a stable identity (e.g., Kafka broker IDs)
- Each instance needs its own isolated storage (e.g., database data directories)
- Startup/shutdown order matters (e.g., primary must start before replicas)
- Examples: MySQL, PostgreSQL, MongoDB, Kafka, Elasticsearch, Cassandra, ZooKeeper

### Headless Service

A headless service is created with `clusterIP: None`. Unlike a regular ClusterIP service that load-balances across pods, a headless service creates individual DNS A records for each pod:

```
<pod-name>.<service-name>.<namespace>.svc.cluster.local
```

This allows direct addressing of individual pods — essential for StatefulSet workloads where you need to target a specific instance (e.g., always write to the primary, read from a specific replica).

In this lab the headless service `devops-python-devops-python-headless` created the following DNS records:
- `devops-python-devops-python-0.devops-python-devops-python-headless.default.svc.cluster.local`
- `devops-python-devops-python-1.devops-python-devops-python-headless.default.svc.cluster.local`
- `devops-python-devops-python-2.devops-python-devops-python-headless.default.svc.cluster.local`

---

## 2. Implementation

### Files Created

- `k8s/devops-python/templates/statefulset.yaml` — StatefulSet with `volumeClaimTemplates`
- `k8s/devops-python/templates/service-headless.yaml` — Headless service (`clusterIP: None`)
- `k8s/devops-python/values.yaml` — Added `statefulset.enabled: true`
- `k8s/devops-python/templates/pvc.yaml` — Gated with `{{- if not .Values.statefulset.enabled }}` to avoid PVC conflict

### Storage Configuration in values.yaml

Storage size and class are fully configurable via `values.yaml`:

```yaml
persistence:
  enabled: true
  size: 100Mi       # passed to volumeClaimTemplates storage request
  storageClass: ""  # empty = use cluster default (standard on minikube)
```

To override at deploy time:
```bash
helm install devops-python k8s/devops-python --set persistence.size=500Mi --set persistence.storageClass=fast
```

### StatefulSet Template Key Sections

```yaml
spec:
  serviceName: devops-python-devops-python-headless  # Links to headless service
  replicas: 3
  volumeClaimTemplates:                              # Per-pod PVC creation
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      {{- if .Values.persistence.storageClass }}
      storageClassName: {{ .Values.persistence.storageClass }}
      {{- end }}
      resources:
        requests:
          storage: {{ .Values.persistence.size }}   # driven from values.yaml
```

### Rendered Kinds (helm template)

```
kind: Secret
kind: ConfigMap
kind: ConfigMap
kind: Service          # Regular ClusterIP for external access
kind: Service          # Headless (clusterIP: None) for pod DNS
kind: StatefulSet
kind: Job              # pre-install hook
kind: Job              # post-install hook
```

---

## 3. Resource Verification

```
NAME                                                 READY   STATUS      RESTARTS   AGE
pod/devops-python-devops-python-0                    1/1     Running     0          2m44s
pod/devops-python-devops-python-1                    1/1     Running     0          119s
pod/devops-python-devops-python-2                    1/1     Running     0          100s
pod/devops-python-devops-python-post-install-j57xg   0/1     Completed   0          2m44s
pod/devops-python-devops-python-pre-install-jz6bt    0/1     Completed   0          3m6s

NAME                                           READY   AGE
statefulset.apps/devops-python-devops-python   3/3     2m44s

NAME                                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/devops-python-devops-python            ClusterIP   10.96.132.128   <none>        80/TCP    2m44s
service/devops-python-devops-python-headless   ClusterIP   None            <none>        80/TCP    2m44s
service/kubernetes                             ClusterIP   10.96.0.1       <none>        443/TCP   11m

NAME                                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-devops-python-devops-python-0   Bound    pvc-cb21212f-1f44-4336-85b7-fb75a918d5c4   100Mi      RWO            standard       2m44s
persistentvolumeclaim/data-devops-python-devops-python-1   Bound    pvc-1c0125ed-d08d-43fc-9a9f-af74071f32bf   100Mi      RWO            standard       119s
persistentvolumeclaim/data-devops-python-devops-python-2   Bound    pvc-8dd480d4-4062-47b8-a5f7-2d832104519d   100Mi      RWO            standard       100s
```

Key observations:
- Pods are named with ordinal suffixes (`-0`, `-1`, `-2`), not random hashes
- Headless service shows `ClusterIP: None`
- Three separate PVCs automatically provisioned by `volumeClaimTemplates`, one per pod
- All PVCs are `Bound` to distinct volumes

### Ordered Startup Evidence

Pods started strictly in order as captured by `kubectl get pods -w`:

```
devops-python-devops-python-0    ContainerCreating → Running (pod-1 not started yet)
devops-python-devops-python-1    ContainerCreating → Running (pod-2 not started yet)
devops-python-devops-python-2    Pending → ContainerCreating → Running
```

---

## 4. Network Identity — DNS Resolution

Executed from inside pod-0 using Python's `socket` module (image has no `nslookup`):

```bash
kubectl exec -it devops-python-devops-python-0 -- python3 -c "
import socket
print('pod-0:', socket.gethostbyname('devops-python-devops-python-0.devops-python-devops-python-headless.default.svc.cluster.local'))
print('pod-1:', socket.gethostbyname('devops-python-devops-python-1.devops-python-devops-python-headless.default.svc.cluster.local'))
print('pod-2:', socket.gethostbyname('devops-python-devops-python-2.devops-python-devops-python-headless.default.svc.cluster.local'))
"
```

Output:
```
pod-0: 10.244.0.5
pod-1: 10.244.0.6
pod-2: 10.244.0.7
```

DNS pattern: `<pod-name>.<headless-service-name>.<namespace>.svc.cluster.local`

Each pod resolves to its own unique IP. The headless service does not load-balance — it exposes each pod's IP directly via DNS.

---

## 5. Per-Pod Storage Evidence

Each pod was accessed individually via `kubectl port-forward` and hit with requests to increment its own visit counter:

| Pod | Port-Forward | Requests Made | Final Visit Count |
|-----|-------------|---------------|-------------------|
| pod-0 | `8080:8000` | 3x `GET /` | **3** |
| pod-1 | `8081:8000` | 1x `GET /` | **1** |
| pod-2 | `8082:8000` | 2x `GET /` | **2** |

Sample output from `/visits` endpoint on each pod:

**pod-0** (hostname: `devops-python-devops-python-0`):
```json
{"visits":3,"timestamp":"2026-03-22T05:25:53.734950+00:00"}
```

**pod-1** (hostname: `devops-python-devops-python-1`):
```json
{"visits":1,"timestamp":"2026-03-22T05:26:20.703711+00:00"}
```

**pod-2** (hostname: `devops-python-devops-python-2`):
```json
{"visits":2,"timestamp":"2026-03-22T05:26:47.870370+00:00"}
```

Each pod maintains a completely isolated counter in its own `/data/visits` file on its own PVC. A shared PVC would have shown all pods reading/writing the same counter.

---

## 6. Persistence Test — Data Survives Pod Deletion

**Before deletion** — read raw file from pod-0:
```bash
kubectl exec devops-python-devops-python-0 -- cat /data/visits
3
```

**Delete pod-0:**
```bash
kubectl delete pod devops-python-devops-python-0
pod "devops-python-devops-python-0" deleted from default namespace
```

**StatefulSet immediately recreated pod-0** (observed via `kubectl get pods -w`):
```
devops-python-devops-python-0   0/1   ContainerCreating   0   0s
devops-python-devops-python-0   1/1   Running             0   17s
```

**After restart** — read raw file from new pod-0 container:
```bash
kubectl exec devops-python-devops-python-0 -- cat /data/visits
3
```

Visit count preserved at `3`. The PVC `data-devops-python-devops-python-0` outlived the pod and was automatically reattached to the replacement pod. This is the fundamental guarantee of StatefulSet per-pod storage.

---

## 7. Bonus — Update Strategies

### Partitioned Rolling Update

A partition value causes the rolling update to only apply to pods with ordinal >= partition. Pods below the partition are protected from the update — useful for staged canary-style rollouts within a StatefulSet.

**Set partition to 2:**
```bash
# patch.json: {"spec":{"updateStrategy":{"type":"RollingUpdate","rollingUpdate":{"partition":2}}}}
kubectl patch statefulset devops-python-devops-python --patch-file patch.json
```

**Trigger image update** (latest → 2026.02.11-89e5033):
```bash
# patch-image.json: {"spec":{"template":{"spec":{"containers":[{"name":"devops-python","image":"3llimi/devops-info-service:2026.02.11-89e5033"}]}}}}
kubectl patch statefulset devops-python-devops-python --patch-file patch-image.json
```

**Result — only pod-2 was updated:**
```
devops-python-devops-python-0   3llimi/devops-info-service:latest             (ordinal 0 < partition 2, untouched)
devops-python-devops-python-1   3llimi/devops-info-service:latest             (ordinal 1 < partition 2, untouched)
devops-python-devops-python-2   3llimi/devops-info-service:2026.02.11-89e5033 (ordinal 2 >= partition 2, updated)
```

Use case: test a new version on the last pod only before rolling it out to all pods.

### OnDelete Strategy

With `OnDelete`, Kubernetes updates the StatefulSet spec but never automatically restarts any pod. Pods only pick up the new spec when they are manually deleted. This gives full manual control over exactly when each pod is updated.

**Switch to OnDelete:**
```bash
# patch-ondel.json: {"spec":{"updateStrategy":{"type":"OnDelete","rollingUpdate":null}}}
kubectl patch statefulset devops-python-devops-python --patch-file patch-ondel.json
```

**Trigger image change** (back to latest):
```bash
kubectl patch statefulset devops-python-devops-python --patch-file patch-image.json
```

**Immediately after patch — no pods updated:**
```
devops-python-devops-python-0   3llimi/devops-info-service:latest
devops-python-devops-python-1   3llimi/devops-info-service:latest
devops-python-devops-python-2   3llimi/devops-info-service:latest
```

**Manually delete pod-1:**
```bash
kubectl delete pod devops-python-devops-python-1
```

**After pod-1 restarts — only pod-1 updated:**
```
devops-python-devops-python-0   3llimi/devops-info-service:latest              (not deleted, not updated)
devops-python-devops-python-1   3llimi/devops-info-service:2026.02.11-89e5033  (deleted manually, picked up new spec)
devops-python-devops-python-2   3llimi/devops-info-service:latest              (not deleted, not updated)
```

**Use case:** maintenance windows where each node of a database cluster must be updated one at a time with manual verification between each step. OnDelete ensures no pod is updated without explicit operator action.

### Strategy Comparison

| Strategy | Update Trigger | Control Level | Use Case |
|----------|---------------|---------------|----------|
| `RollingUpdate` (default) | Automatic, ordered | Low — Kubernetes decides timing | Standard updates with no manual intervention needed |
| `RollingUpdate` + partition | Automatic for ordinal >= N | Medium — protect lower pods | Staged rollout, canary testing on last pod |
| `OnDelete` | Manual pod deletion only | High — operator decides per pod | Database maintenance, strict change control |