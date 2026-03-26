# Lab 18 — Reproducible Builds with Nix

- **Environment:** Windows + WSL2 (Ubuntu), Docker Desktop
- **Repository:** `DevOps-Core-Course`
- **Branch used:** `lab18`

---

## 1) Task 1 — Build Reproducible Python App (6 pts)

### 1.1 Nix installation and verification

I installed Nix (Determinate installer) and verified it:

```bash
nix --version
nix run nixpkgs#hello
```

Observed:
- `nix (Determinate Nix 3.17.1) 2.33.3`
- `Hello, world!`

![Nix install and verification](../docs/screenshots/S18-01-nix-install-verify.png)

---

### 1.2 Application preparation

I used the Lab 1/2 Python app in:

- `labs/lab18/app_python/`

The app is FastAPI-based and exposes `/health`.

---

### 1.3 Nix derivation (`default.nix`)

I created `labs/lab18/app_python/default.nix` to package and run the app reproducibly.

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
    annotated-doc
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

Why this derivation:
- Uses Nix-provided Python + dependencies.
- Wraps executable consistently.
- Filters mutable source files for stable output hashes.

---

### 1.4 Build and run

Commands:

```bash
cd labs/lab18/app_python
nix-build
readlink result
./result/bin/devops-info-service
```

Health check:

```bash
curl -s http://localhost:8000/health
```

Example response:

```json
{"status":"healthy","timestamp":"2026-03-26T05:21:29.528356+00:00","uptime_seconds":20}
```

![Task1 app running from nix build](../docs/screenshots/S18-02-task1-nix-run.png)

---

### 1.5 Reproducibility proof

I rebuilt and compared store paths:

```bash
rm -f result
nix-build
readlink result
rm -f result
nix-build
readlink result
```

Both `readlink result` outputs were identical (after source cleanup).

Hash:

```bash
nix-hash --type sha256 result
```

Observed:
- `d4ad3501ab1afad0104576d6e84704971daac215df5e643d7e86927e44235658`

![Task1 reproducibility proof](../docs/screenshots/S18-03-task1-reproducible-storepath.png)

---

### 1.6 Comparison with traditional pip workflow

Traditional `venv + pip install -r requirements.txt` is weaker because:
- depends on host runtime,
- transitive dependencies may drift,
- reproducibility over time is weaker.

Nix is stronger because:
- full dependency graph and build inputs are explicit,
- outputs are content-addressed in `/nix/store`,
- same inputs produce same output path/hash.

---

## 2) Task 2 — Reproducible Docker Images with Nix (4 pts)

### 2.1 Nix Docker expression (`docker.nix`)

I created `labs/lab18/app_python/docker.nix`:

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

---

### 2.2 Build image tarball with Nix

Commands:

```bash
cd labs/lab18/app_python
nix-build docker.nix
readlink result
```

Example output tarball:
- `/nix/store/2lmnk34d6hd1brq3fnpkril8va0dzgnv-devops-info-service-nix.tar.gz`

---

### 2.3 Load image into Docker

I loaded image in PowerShell from WSL path:

```powershell
docker load -i "\\wsl$\Ubuntu\nix\store\2lmnk34d6hd1brq3fnpkril8va0dzgnv-devops-info-service-nix.tar.gz"
docker images | findstr devops-info-service-nix
```

Observed:
- `Loaded image: devops-info-service-nix:1.0.0`

![Task2 nix docker build and load](../docs/screenshots/S18-04-task2-nix-docker-build-load.png)

---

### 2.4 Run side-by-side with Lab 2 Docker image

Commands:

```powershell
docker rm -f lab2-container nix-container 2>$null
docker run -d -p 5000:8000 --name lab2-container lab2-app:v1
docker run -d -p 5001:8000 --name nix-container devops-info-service-nix:1.0.0
curl.exe -s http://localhost:5000/health
curl.exe -s http://localhost:5001/health
docker ps
```

Both health checks returned successful JSON and both containers were running.

![Task2 both containers healthy](../docs/screenshots/S18-05-task2-both-containers-health.png)

---

### 2.5 Analysis: Dockerfile vs Nix dockerTools

| Aspect | Traditional Dockerfile | Nix dockerTools |
|---|---|---|
| Dependency source | Base image + runtime install | Nix store closure |
| Determinism | Weaker (tags/metadata/time effects) | Stronger (content-addressed derivations) |
| Build repeatability | Can vary | Highly stable for fixed inputs |
| Traceability | Layer-oriented | Full dependency closure in Nix store |

Traditional Dockerfiles are practical but usually not bit-for-bit reproducible by default. Nix gives stronger reproducibility guarantees by fully controlling build inputs.

---

## 3) Challenges and fixes

1. **Script execution issue (`from: command not found`)**  
   Fixed by wrapping app with explicit Python interpreter via `makeWrapper`.

2. **Missing module (`annotated_doc`)**  
   Fixed by adding `annotated-doc` to Python environment.

3. **Changing store paths across rebuilds**  
   Fixed with `cleanSourceWith` to remove mutable files from build input.

4. **Docker CLI issue in WSL session**  
   Loaded Nix tar from PowerShell via `\\wsl$\...` path.

---

## 4) Reflection

Using Nix from the start (Lab 1/2) would have improved consistency, reduced environment drift, and made CI/CD artifacts more deterministic and auditable. Docker remains useful for runtime packaging; Nix strengthens reproducible builds.

---