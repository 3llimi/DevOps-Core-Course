# Lab 13 - GitOps with ArgoCD

## 1. ArgoCD Setup

### Installation
- Added Helm repo: `argo`
- Installed ArgoCD into `argocd` namespace using Helm chart `argo/argo-cd`
- Verified all core components are running:
  - argocd-server
  - argocd-repo-server
  - argocd-application-controller
  - argocd-applicationset-controller
  - argocd-dex-server
  - argocd-redis

### UI Access
- Port-forward method used:

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

- UI endpoint: https://localhost:8080
- Initial admin password retrieved from `argocd-initial-admin-secret`

### CLI Access
- Installed ArgoCD CLI (winget)
- Logged in with:

```bash
argocd login localhost:8080 --insecure --username admin --password <initial-password>
```

- Verified with:

```bash
argocd version
argocd app list
```

## 2. Application Configuration

Created ArgoCD application manifests in `k8s/argocd/`:
- `application.yaml` (default namespace, manual sync)
- `application-dev.yaml` (dev namespace, auto-sync + prune + selfHeal)
- `application-prod.yaml` (prod namespace, manual sync)

Source and destination configuration:
- Repo: `https://github.com/3llimi/DevOps-Core-Course.git`
- Revision: `lab13`
- Chart path: `k8s/devops-python`
- Helm value files:
  - default: `values.yaml`
  - dev: `values-dev.yaml`
  - prod: `values-prod.yaml`

## 3. Multi-Environment Deployment

### Namespaces
- Created:
  - `dev`
  - `prod`

### Environment differences
- Dev uses `values-dev.yaml`:
  - `replicaCount: 1`
  - lower CPU/memory requests and limits
  - `nodePort: 30081`
  - auto-sync enabled
- Prod uses `values-prod.yaml`:
  - `replicaCount: 5`
  - higher CPU/memory requests and limits
  - `nodePort: 30082`
  - manual sync (approval gate)

### Why prod is manual
- Controlled release timing
- Change review before deployment
- Lower risk for production incidents
- Better rollback readiness

## 4. Self-Healing and Sync Behavior

### Manual scale drift test (ArgoCD self-heal)
- Scaled dev deployment manually with kubectl (drift from Git desired state)
- ArgoCD detected OutOfSync and reconciled back to Git-defined replica count
- Result: drift reverted automatically in dev due to `selfHeal: true`

### Pod deletion test (Kubernetes self-heal)
- Deleted a dev pod manually
- ReplicaSet/Deployment recreated pod automatically
- Result: this is Kubernetes controller behavior, not ArgoCD reconciliation

### Config drift test
- Changed an in-cluster resource directly (outside Git)
- ArgoCD detected state mismatch and reconciled to Git state in dev
- Result: Git remained source of truth

### Sync triggers and intervals
- Manual sync via `argocd app sync <app>`
- Auto-sync for apps with automated policy enabled
- Git polling is periodic (typically ~3 minutes by default), with optional webhook acceleration

## 5. GitOps Workflow Evidence

Performed workflow:
1. Changed chart configuration in Git (`replicaCount`)
2. Committed and pushed to `lab13`
3. ArgoCD showed OutOfSync
4. Synced app and cluster converged to Git state

Important fix applied:
- Set `targetRevision: lab13` in application manifests to track the active branch
- Avoided NodePort conflict by assigning unique ports per environment (`30081`, `30082`)

## 6. Current State Summary

Applications visible in ArgoCD:
- `devops-python` (default, manual) - Synced/Healthy
- `devops-python-dev` (dev, auto) - Synced/Healthy
- `devops-python-prod` (prod, manual) - Synced/Healthy

Environment pods:
- `dev`: running with dev profile
- `prod`: running with prod profile
- `default`: running with default profile

## 7. Screenshots to Include

Add these screenshots to your submission:
1. ArgoCD Applications list showing all 3 apps
2. Application details page (source/destination/sync policy)
3. Dev app auto-sync/self-heal evidence
4. Dev vs prod namespace workloads (`kubectl get pods -n dev/prod`)

## Bonus - ApplicationSet

Implemented `k8s/argocd/applicationset.yaml` with List generator for dev/prod.

Benefits of ApplicationSet:
- One template generates multiple applications
- Less duplication than separate Application manifests
- Easier scaling to many environments
- Consistent naming and structure

When to use which:
- Individual Application: small setup, explicit control
- ApplicationSet: multiple similar environments/apps, scalable GitOps management
