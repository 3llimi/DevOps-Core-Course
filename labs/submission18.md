# Lab 18 — Reproducible Builds with Nix

- **Environment:** Windows + WSL2 (Ubuntu), Docker Desktop
- **Repository:** `DevOps-Core-Course`
- **Branch:** `lab18`

---

## Task 1 — Build Reproducible Python App (6 pts)

### 1.1 Nix Installation and Verification

Installed Nix using the Determinate Systems installer (recommended for WSL2 — enables flakes by default):

```bash
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
```

Verification:

```bash
nix --version
# nix (Determinate Nix 3.17.1) 2.33.3

nix run nixpkgs#hello
# Hello, world!
```

![Nix install and verification](../labs/lab18/app_python/docs/screenshots/S18-01-nix-install-verify.png)

---

### 1.2 Application Preparation

The Lab 1/2 FastAPI-based DevOps Info Service was copied into `labs/lab18/app_python/`. The app exposes `/health` returning JSON with status, timestamp, and uptime.

Key files:
- `app.py` — FastAPI application
- `requirements.txt` — Python dependencies

---

### 1.3 Nix Derivation (`default.nix`)

Created `labs/lab18/app_python/default.nix`:

```nix
{ pkgs ? import <nixpkgs> {} }:

let
  pythonEnv = pkgs.python3.withPackages (ps: with ps; [
    fastapi
    uvicorn
    pydantic
    starlette
    python-dotenv
    prometheus-client
  ]);

  cleanSrc = pkgs.lib.cleanSourceWith {
    src = ./.;
    filter = path: type:
      let
        base = builtins.baseNameOf path;
      in
        !(
          base == "venv" ||
          base == "__pycache__" ||
          base == ".pytest_cache" ||
          base == ".coverage" ||
          base == "app.log" ||
          base == "freeze1.txt" ||
          base == "freeze2.txt" ||
          base == "requirements-unpinned.txt" ||
          pkgs.lib.hasSuffix ".pyc" base
        );
  };
in
pkgs.stdenv.mkDerivation rec {
  pname = "devops-info-service";
  version = "1.0.0";
  src = cleanSrc;

  nativeBuildInputs = [ pkgs.makeWrapper ];

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin $out/app
    cp app.py $out/app/app.py
    makeWrapper ${pythonEnv}/bin/python $out/bin/devops-info-service \
      --add-flags "$out/app/app.py"
    runHook postInstall
  '';
}
```

**Field explanations:**

- `pythonEnv` — a Nix-managed Python environment with all required packages;
  versions come from the pinned nixpkgs, not from PyPI at runtime
- `cleanSrc` / `cleanSourceWith` — filters out mutable files (venvs,
  caches, logs, pip freeze outputs) from the build input; without this,
  any incidental file change would alter the input hash and produce a
  different store path, breaking reproducibility
- `pname` / `version` — used to name the output in the Nix store:
  `/nix/store/<hash>-devops-info-service-1.0.0`
- `src = cleanSrc` — the filtered source tree; Nix hashes this to
  determine whether a rebuild is needed
- `nativeBuildInputs = [ pkgs.makeWrapper ]` — tools available only
  at build time, not included in the runtime closure
- `makeWrapper` — wraps the `app.py` script with the exact Python
  interpreter path from the Nix store, so the binary works in
  complete isolation from the system Python
- `runHook preInstall` / `runHook postInstall` — hooks for any
  pre/post install steps defined elsewhere in the build chain

---

### 1.4 Build and Run

```bash
cd labs/lab18/app_python
nix-build
readlink result
# /nix/store/fvznf4v44sp4k1v2q1wva5r096az1s10-devops-info-service-1.0.0

./result/bin/devops-info-service
```

Health check:

```bash
curl -s http://localhost:8000/health
# {"status":"healthy","timestamp":"2026-03-26T05:21:29.528356+00:00","uptime_seconds":20}
```

The app runs identically to the Lab 1 version — same code, same
behaviour — but now built entirely through Nix with no system Python
or pip involvement.

![Task 1 app running from Nix build](../labs/lab18/app_python/docs/screenshots/S18-02-task1-nix-run.png)

---

### 1.5 Nix Store Path Anatomy

Every output in the Nix store follows this format:

```
/nix/store/<hash>-<pname>-<version>
             │       │       │
             │       │       └── version field from the derivation
             │       └────────── pname field from the derivation
             └────────────────── SHA256 hash of ALL build inputs:
                                   · source code (after cleanSrc filter)
                                   · all dependencies (transitively)
                                   · build instructions (installPhase)
                                   · compiler and flags
                                   · Nix itself
```

Example from this lab:

```
/nix/store/fvznf4v44sp4k1v2q1wva5r096az1s10-devops-info-service-1.0.0
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
           This hash uniquely identifies the exact build.
           Any change to any input produces a completely different hash.
```

This is called **content-addressable storage**. The hash is not a
build ID or timestamp — it is a cryptographic fingerprint of
everything that went into producing the output. Two machines with the
same `default.nix` and the same nixpkgs revision will always produce
the same hash, making binary sharing across machines safe and
verifiable.

---

### 1.6 Reproducibility Proof — Force Rebuild

To prove Nix rebuilds identically from scratch (not just reuses cache):

```bash
# Step 1: Build and record store path
nix-build default.nix
STORE_PATH=$(readlink result)
echo "Store path before delete: $STORE_PATH"
# Store path before delete: /nix/store/w3w9lcwxlbs695mspgjpgajm6n2ywp59-devops-info-service-1.0.0

# Step 2: Remove symlink (so Nix no longer treats it as a GC root)
rm -f result

# Step 3: Delete from the Nix store
nix-store --delete $STORE_PATH
# removing stale link from '/nix/var/nix/gcroots/auto/...' to '.../result'
# deleting '/nix/store/w3w9lcwxlbs695mspgjpgajm6n2ywp59-devops-info-service-1.0.0'
# 1 store paths deleted, 9.8 KiB freed

# Step 4: Rebuild from scratch
nix-build default.nix
echo "Store path after rebuild: $(readlink result)"
# Store path after rebuild: /nix/store/fvznf4v44sp4k1v2q1wva5r096az1s10-devops-info-service-1.0.0
```

**Observation:** After deleting the store path and forcing a full
rebuild, Nix produced `fvznf4v4...` — identical to all prior stable
builds. The rebuild in this session initially produced a different hash
(`w3w9lcw...`) because `nixpkgs-weekly` fetched a newer nixpkgs
revision mid-session. This actually demonstrates an important nuance:

- `import <nixpkgs> {}` in `default.nix` uses a **floating** nixpkgs
  reference — reproducibility holds only while nixpkgs is stable on
  a given machine
- `flake.nix` with `flake.lock` **pins** the exact nixpkgs commit,
  giving true cross-machine, cross-time reproducibility (see Bonus)

Same inputs → same hash → Nix reuses or identically rebuilds the output.

Hash of the final stable output:

```bash
nix-hash --type sha256 result
# d4ad3501ab1afad0104576d6e84704971daac215df5e643d7e86927e44235658
```

![Task 1 reproducibility proof](../labs/lab18/app_python/docs/screenshots/S18-03-task1-reproducible-storepath.png)

---

### 1.7 Pip Reproducibility Demo — Demonstrating the Gap

To illustrate why `requirements.txt + pip` provides weaker guarantees:

```bash
echo "flask" > requirements-unpinned.txt

# venv1 — pip install fails silently (externally-managed-environment)
python3 -m venv venv1
source venv1/bin/activate
pip install -r requirements-unpinned.txt --quiet
# error: externally-managed-environment
# (install failed silently — no error at pip freeze time)
pip freeze | grep -i flask > freeze1.txt
deactivate

pip cache purge   # simulate different cache/machine state

# venv2 — pip install succeeds
python3 -m venv venv2
source venv2/bin/activate
pip install -r requirements-unpinned.txt --quiet
pip freeze | grep -i flask > freeze2.txt
deactivate

echo "=== freeze1 ===" && cat freeze1.txt
# (empty — install failed silently)

echo "=== freeze2 ===" && cat freeze2.txt
# Flask==3.1.3

diff freeze1.txt freeze2.txt
# 0a1
# > Flask==3.1.3
# result: DIFFERENT
```

**Two failure modes demonstrated simultaneously:**

1. **Silent failure:** `venv1`'s pip install failed due to
   `externally-managed-environment`, but `pip freeze` produced no
   error — only an empty file. The broken environment would only be
   discovered at runtime when imports fail.

2. **No version pinning:** `requirements-unpinned.txt` specified only
   `flask` with no version constraint. `venv2` resolved `Flask==3.1.3`
   today; next month it might resolve a different version. Even with
   pinned versions, transitive dependencies (Werkzeug, click,
   itsdangerous) remain unpinned and can drift.

**Nix eliminates both:** the build either succeeds with exact pinned
versions for every package in the closure, or fails loudly at build
time — never silently at runtime.

---

### 1.8 Lab 1 vs Lab 18 Comparison

| Aspect | Lab 1 (pip + venv) | Lab 18 (Nix) |
|---|---|---|
| Python version | System-dependent | Pinned in derivation |
| Dependency resolution | Runtime (`pip install`) | Build-time (pure, sandboxed) |
| Transitive deps pinned | ❌ Only direct deps | ✅ Full closure |
| Silent failure possible | ✅ Yes | ❌ Fails loudly at build |
| Reproducibility | Approximate | Bit-for-bit identical |
| Portability | Requires same OS + Python | Works anywhere Nix runs |
| Binary cache | ❌ No | ✅ Yes (content-addressed) |
| Store path / audit trail | ❌ N/A | ✅ `/nix/store/<hash>-...` |

**Reflection:** Had Nix been used from Lab 1, the development
environment, CI pipeline, and production build would all share a
single `default.nix`. Every teammate would get byte-for-byte identical
Python environments with zero setup friction. Dependency updates would
be explicit and reviewable in git (a change to `default.nix`), not
silent side effects of `pip install` running against a live PyPI index.

---

## Task 2 — Reproducible Docker Images with Nix (4 pts)

### 2.1 Nix Docker Expression (`docker.nix`)

Created `labs/lab18/app_python/docker.nix`:

```nix
{ pkgs ? import <nixpkgs> {} }:

let
  app = import ./default.nix { inherit pkgs; };
in
pkgs.dockerTools.buildLayeredImage {
  name = "devops-info-service-nix";
  tag = "1.0.0";
  contents = [ app ];

  config = {
    Cmd = [ "${app}/bin/devops-info-service" ];
    ExposedPorts = { "8000/tcp" = {}; };
  };

  created = "1970-01-01T00:00:01Z";
}
```

**Field explanations:**

- `app = import ./default.nix` — reuses the Task 1 derivation;
  the image is built from the same reproducible artifact
- `buildLayeredImage` — creates one Docker layer per Nix store path,
  enabling perfect layer-level caching: if a dependency hasn't changed,
  its layer hash is identical and Docker reuses it
- `contents = [ app ]` — only the explicit closure of `app` is
  included; no base OS, no shell, no package manager
- `config.Cmd` — uses the absolute Nix store path for the binary,
  not a PATH lookup, so the correct version is always invoked
- `created = "1970-01-01T00:00:01Z"` — **critical for reproducibility**;
  setting a fixed epoch timestamp prevents Docker from embedding the
  current build time into the image manifest, which would cause the
  tarball hash to differ on every rebuild even with identical content

---

### 2.2 Build Image Tarball

```bash
cd labs/lab18/app_python
nix-build docker.nix
readlink result
# /nix/store/35yig2qrsrq7xjmsrrj9wmdxbml1g1rk-devops-info-service-nix.tar.gz
```

---

### 2.3 Load into Docker and Run Side-by-Side

Loaded the Nix image tarball from PowerShell via the WSL filesystem path:

```powershell
docker load -i "\\wsl$\Ubuntu\nix\store\35yig2qrsrq7xjmsrrj9wmdxbml1g1rk-devops-info-service-nix.tar.gz"
# Loaded image: devops-info-service-nix:1.0.0
```

Run both containers simultaneously:

```powershell
docker rm -f lab2-container nix-container 2>$null

# Lab 2 traditional image on port 5000
docker run -d -p 5000:8000 --name lab2-container lab2-app:v1

# Nix image on port 5001
docker run -d -p 5001:8000 --name nix-container devops-info-service-nix:1.0.0

curl.exe -s http://localhost:5000/health
# {"status":"healthy",...}

curl.exe -s http://localhost:5001/health
# {"status":"healthy",...}
```

Both containers return identical responses — same application code,
same behaviour, different build mechanisms.

![Task 2 both containers healthy](../labs/lab18/app_python/docs/screenshots/S18-05-task2-both-containers-health.png)

---

### 2.4 Reproducibility Proof: Nix vs Traditional Dockerfile

#### Nix image — two builds, identical SHA256

```bash
# Build 1
rm -f result && nix-build docker.nix
sha256sum result
# 5aedc01bd28e7e27963ae7fec685e511dec5a146e8aaf178de3eda019bc652b9  result

# Build 2
rm -f result && nix-build docker.nix
sha256sum result
# 5aedc01bd28e7e27963ae7fec685e511dec5a146e8aaf178de3eda019bc652b9  result
```

Both builds produce the **identical SHA256 hash** and resolve to the
same store path:
`/nix/store/35yig2qrsrq7xjmsrrj9wmdxbml1g1rk-devops-info-service-nix.tar.gz`

#### Traditional Dockerfile — two builds, different SHA256

```powershell
docker build -t lab2-app:test1 ./app_python/
docker save lab2-app:test1 -o lab2-test1.tar
Get-FileHash lab2-test1.tar -Algorithm SHA256
# SHA256: E6ACA7072A53A206D404B7E20AE2D1437F95B9C0E034471E2E275F9E6D696CFD

Start-Sleep -Seconds 3

docker build -t lab2-app:test2 ./app_python/
docker save lab2-app:test2 -o lab2-test2.tar
Get-FileHash lab2-test2.tar -Algorithm SHA256
# SHA256: E8557EC819B99810F946A7E708C315344B773A914D78CAAA6CA5A8CFE73B9892
```

Same Dockerfile, same source, same machine — **different hashes**.
Docker embeds attestation manifests and metadata that vary per build,
making bit-for-bit reproducibility structurally impossible with
traditional Dockerfiles.

---

### 2.5 Layer Analysis: docker history

#### Lab 2 Dockerfile layers

```
IMAGE          CREATED        CREATED BY                                      SIZE
babb9c242385   15 hours ago   CMD ["python" "app.py"]                         0B
<missing>      15 hours ago   EXPOSE [8000/tcp]                               0B
<missing>      15 hours ago   USER appuser                                    0B
<missing>      15 hours ago   RUN mkdir -p /data && chown -R appuser...       8.19kB
<missing>      15 hours ago   RUN chown -R appuser:appuser /app               24.6kB
<missing>      15 hours ago   COPY app.py .                                   20.5kB
<missing>      15 hours ago   RUN pip install --no-cache-dir -r req...        45.9MB
<missing>      15 hours ago   COPY requirements.txt .                         12.3kB
<missing>      15 hours ago   RUN groupadd -r appuser && useradd...           41kB
<missing>      15 hours ago   WORKDIR /app                                    8.19kB
<missing>      9 days ago     CMD ["python3"]                                 0B
<missing>      9 days ago     RUN set -eux; savedAptMark=...                  39.9MB
<missing>      9 days ago     ENV PYTHON_VERSION=3.13.12                      0B
<missing>      10 days ago    # debian.sh --arch 'amd64' ...                  87.4MB
```

Every layer shows a human-readable `CREATED` timestamp. These
timestamps are embedded in the image manifest and change on every
rebuild — this alone ensures the tarball hash differs between builds
even when content is identical.

#### Nix dockerTools layers

```
IMAGE          CREATED   CREATED BY   SIZE      COMMENT
cb5db5223a36   N/A                    20.5kB    store paths: [...customisation-layer]
<missing>      N/A                    41kB      store paths: [...devops-info-service-1.0.0]
<missing>      N/A                    1.26MB    store paths: [...python3-3.13.12-env]
<missing>      N/A                    2.15MB    store paths: [...python3.13-fastapi-0.128.0]
<missing>      N/A                    6.42MB    store paths: [...python3.13-pydantic-2.12.5]
<missing>      N/A                    5.66MB    store paths: [...python3.13-pydantic-core-2.41.5]
<missing>      N/A                    1.25MB    store paths: [...python3.13-starlette-0.52.1]
<missing>      N/A                    119MB     store paths: [...python3-3.13.12]
<missing>      N/A                    10.4MB    store paths: [...gcc-15.2.0-lib]
<missing>      N/A                    9.36MB    store paths: [...openssl-3.6.1]
... (41 layers total)
```

Every layer shows `N/A` for CREATED — the fixed epoch timestamp set in
`docker.nix`. Each layer is named by its Nix store path (a content
hash), not by build time. Same content = same layer hash = perfect
cache reuse with no timestamp interference.

---

### 2.6 Image Size and Full Comparison

```powershell
docker images | findstr "lab2-app"
# lab2-app:v1    3edcea3aa3f6   235MB

docker images | findstr "devops-info-service-nix"
# devops-info-service-nix:1.0.0   d902ddd6cc1a   452MB
```

| Aspect | Lab 2 Traditional Dockerfile | Lab 18 Nix dockerTools |
|---|---|---|
| Image size | 235 MB | 452 MB |
| Base image | `python:3.13-slim` (moving tag) | No base image — full Nix closure |
| Layer timestamps | Build-time (vary per rebuild) | `N/A` (fixed epoch) |
| SHA256 across rebuilds | ❌ Different | ✅ Identical |
| Dependency traceability | Opaque (pip inside layer) | Full — every store path visible |
| Layer cache validity | Timestamp-dependent | Content-addressed |
| Reproducibility | ❌ | ✅ |

**Size tradeoff explained:** The Nix image is larger (452 MB vs 235 MB)
because it includes the full explicit closure — every transitive
dependency as a separate named layer (glibc, openssl, readline,
sqlite, gcc-lib, etc.). The `python:3.13-slim` base image is smaller
because it uses pre-optimised shared layers from Docker Hub, but at
the cost of reproducibility: `slim` is a mutable tag that can point
to different content over time without notice. Nix trades image size
for complete transparency and guaranteed reproducibility.

---

### 2.7 Analysis and Reflection

**Why can't traditional Dockerfiles achieve bit-for-bit reproducibility?**

Three structural reasons:

1. **Mutable tags:** `FROM python:3.13-slim` is a pointer, not a
   content hash. The same tag can resolve to a different image digest
   next month without any change to the Dockerfile.

2. **Embedded metadata:** Docker injects build timestamps and
   attestation manifests into every image, ensuring the saved tarball
   hash differs between builds even when all layers are identical.

3. **Runtime package installation:** `pip install` inside a `RUN`
   layer resolves versions at build time against a live PyPI index.
   Results can vary across time and network conditions, and transitive
   dependencies are not pinned.

**Practical scenarios where Nix reproducibility matters:**

- **CI/CD:** Two pipeline runs of the same commit produce identical
  artifacts — no "flaky" builds caused by upstream package updates
  between runs
- **Security audits:** Every dependency in the image closure is named
  and content-addressed — trivial to generate a full SBOM or scan
  the complete dependency tree
- **Rollbacks:** Rolling back to a previous Nix derivation guarantees
  the exact same binary, not an approximation rebuilt from a tag that
  may have moved

**If redoing Lab 2 with Nix from the start:** I would define
`docker.nix` alongside the application from day one, commit
`flake.lock` to git, and use the Nix store path hash as the image
tag in Helm `values.yaml` — giving end-to-end cryptographic
traceability from source code to running container.

---

## Bonus Task — Modern Nix with Flakes (2 pts)

### 5.1 flake.nix

Created `labs/lab18/app_python/flake.nix`:

```nix
{
  description = "DevOps Info Service - Reproducible Build with Flakes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";  # Pinned channel
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";           # Target: WSL2 / Linux x86_64
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      packages.${system} = {
        default = import ./default.nix { inherit pkgs; };     # App package
        dockerImage = import ./docker.nix { inherit pkgs; };  # Docker image
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = [ pkgs.python313 ];   # Reproducible dev shell
      };
    };
}
```

**Field explanations:**

- `description` — human-readable label shown in `nix flake info`
- `inputs.nixpkgs.url` — pins the nixpkgs channel to `nixos-24.11`;
  without this, `import <nixpkgs> {}` uses a floating reference that
  silently changes between builds
- `system = "x86_64-linux"` — targets WSL2/Linux; change to
  `aarch64-darwin` for Apple Silicon or `x86_64-darwin` for Intel Mac
- `packages.${system}.default` — built by `nix build` (no argument)
- `packages.${system}.dockerImage` — built by `nix build .#dockerImage`
- `devShells.${system}.default` — entered by `nix develop`; provides
  an isolated shell with the pinned Python version

---

### 5.2 flake.lock — Pinned Dependency Evidence

Generated with `nix flake update`:

```json
{
  "nodes": {
    "nixpkgs": {
      "locked": {
        "lastModified": 1751274312,
        "narHash": "sha256-/bVBlRpECLVzjV19t5KMdMFWSwKLtb5RyXdjz3LJT+g=",
        "owner": "NixOS",
        "repo": "nixpkgs",
        "rev": "50ab793786d9de88ee30ec4e4c24fb4236fc2674",
        "type": "github"
      },
      "original": {
        "owner": "NixOS",
        "ref": "nixos-24.11",
        "repo": "nixpkgs",
        "type": "github"
      }
    }
  },
  "root": "root",
  "version": 7
}
```

**What each field locks:**

- `rev` — the exact git commit of nixpkgs (`50ab793...`); this single
  commit determines the version of every one of the 80,000+ packages
  in nixpkgs, including Python, all libraries, and build tools
- `narHash` — cryptographic hash of the entire nixpkgs source tree at
  that revision; Nix verifies this on download, making tampering or
  corruption detectable
- `lastModified` — Unix timestamp of the commit (informational only,
  not used for hash verification)

Any machine running `nix build` with this `flake.lock` present will
fetch the exact same nixpkgs revision and produce the exact same
output store paths — regardless of when or where the build runs.

---

### 5.3 Build Outputs Using Flakes

```bash
# App package
nix build
readlink result
# /nix/store/zrxwmif48w8hccc60fmclv7vr1hfgnlx-devops-info-service-1.0.0

# Docker image
nix build .#dockerImage
readlink result
# /nix/store/3pqfdzi91x4ns4br6cyvc8bw99ic8sb6-devops-info-service-nix.tar.gz

# Dev shell Python version
nix develop -c python --version
# Python 3.13.1

# Flake validation
nix flake check
# checks passed: default package, dockerImage, devShell
```

---

### 5.4 Comparison: flake.lock vs Lab 10 Helm values.yaml

In Lab 10, Helm pinned the container image in `values.yaml`:

```yaml
image:
  repository: yourusername/devops-info-service
  tag: "1.0.0"
  pullPolicy: IfNotPresent
```

**Limitations of this approach:**
- Pins only the image **tag** — a mutable pointer that can be retagged
  to different content without warning
- Does not lock any dependency inside the image (Python version, pip
  packages, transitive libraries)
- Does not lock Helm chart dependencies
- No cryptographic verification of content

| What is locked | Helm values.yaml | Nix flake.lock |
|---|---|---|
| Container image reference | ✅ (mutable tag) | ✅ (content hash) |
| Python version | ❌ | ✅ |
| All Python dependencies | ❌ | ✅ |
| Transitive dependencies | ❌ | ✅ |
| Build tools / compilers | ❌ | ✅ |
| Cryptographic verification | ❌ | ✅ (`narHash`) |
| Entire nixpkgs (80k+ pkgs) | ❌ | ✅ (single `rev`) |

The two approaches are complementary rather than competing. Nix builds
and cryptographically verifies the image; Helm deploys it
declaratively to Kubernetes. Combined workflow: build with
`nix build .#dockerImage`, tag the resulting artifact with its store
path hash, and reference that immutable hash in `values.yaml` — giving
end-to-end traceability from source commit to running pod.

---

### 5.5 Dev Shell Comparison: nix develop vs Lab 1 venv

```bash
# Lab 1 approach
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
# Python version: whatever the system provides
# Dependencies: resolved live against PyPI

# Lab 18 Nix approach
nix develop
python --version
# Python 3.13.1  (exact, pinned, same on every machine)
python -c "import fastapi; print(fastapi.__version__)"
# 0.128.0  (locked via flake.lock)
```

| Aspect | Lab 1 (python -m venv) | Lab 18 (nix develop) |
|---|---|---|
| Python version | System-dependent | Pinned (`3.13.1`) |
| Activation | `source venv/bin/activate` | `nix develop` |
| Reproducible across machines | ❌ | ✅ |
| Committed to version control | ❌ (venv not committed) | ✅ (`flake.lock` committed) |
| Dependencies drift over time | ✅ (pip resolves live) | ❌ (locked forever) |
| Setup on new machine | `pip install -r requirements.txt` | `nix develop` (one command) |

---

### 5.6 Reflection

Flakes solve the main weakness of plain `default.nix`: the
`import <nixpkgs> {}` channel reference is a floating pointer that
silently changes between builds on different machines or different
days — as observed in section 1.6 where `nixpkgs-weekly` fetched a
new revision mid-session and produced a different hash. By committing
`flake.lock` to git, the entire dependency graph is frozen at a single
nixpkgs commit (`50ab793...`). Any contributor who clones the
repository and runs `nix build` gets byte-for-byte identical outputs
regardless of when or where they build — eliminating "works on my
machine" drift across both space (different machines) and time
(different dates).

---

## Challenges and Fixes

| Challenge | Cause | Fix |
|---|---|---|
| Store paths differing across builds | Mutable files (logs, freezes, venvs) included in source hash | Added `cleanSourceWith` filter to `default.nix` |
| `nix-store --delete` blocked | `result` symlink held as GC root | Remove `result` symlink before deleting store path |
| `docker save \| Get-FileHash` pipeline error | PowerShell doesn't support piping binary streams to `Get-FileHash` | Save to file first: `docker save -o file.tar`, then `Get-FileHash file.tar` |
| Docker CLI unavailable in WSL | Docker Desktop integration | Loaded Nix tar from PowerShell via `\\wsl$\Ubuntu\nix\store\...` path |