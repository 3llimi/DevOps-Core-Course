# Lab 05 — Ansible Fundamentals

## 1. Architecture Overview

**Ansible Version:** 2.10.8
**Target VM OS:** Ubuntu 22.04 LTS (jammy64)
**Control Node:** Same VM (Ansible runs on the VM and targets itself via `ansible_connection=local`)

### Role Structure

```
ansible/
├── inventory/
│   ├── hosts.ini              # Static inventory (localhost)
│   └── dynamic_inventory.py  # Dynamic inventory script (bonus)
├── roles/
│   ├── common/                # Common system packages
│   │   ├── tasks/main.yml
│   │   └── defaults/main.yml
│   ├── docker/                # Docker installation
│   │   ├── tasks/main.yml
│   │   ├── handlers/main.yml
│   │   └── defaults/main.yml
│   └── app_deploy/            # Application deployment
│       ├── tasks/main.yml
│       ├── handlers/main.yml
│       └── defaults/main.yml
├── playbooks/
│   ├── site.yml               # Main playbook
│   ├── provision.yml          # System provisioning
│   └── deploy.yml             # App deployment
├── group_vars/
│   └── all.yml                # Encrypted variables (Vault)
├── ansible.cfg                # Ansible configuration
└── docs/
    └── LAB05.md
```

### Why Roles Instead of Monolithic Playbooks?

Roles enforce separation of concerns — each role has a single responsibility (common packages, Docker setup, app deployment). This makes the codebase reusable across projects, easier to test independently, and simple to maintain. A monolithic playbook mixing all tasks together would become unmanageable as complexity grows.

---

## 2. Roles Documentation

### common

**Purpose:** Ensures every server has essential system tools installed and the apt cache is up to date.

**Variables (defaults/main.yml):**
```yaml
common_packages:
  - python3-pip
  - curl
  - git
  - vim
  - htop
  - wget
  - unzip
```

**Handlers:** None — package installation does not require service restarts.

**Dependencies:** None.

---

### docker

**Purpose:** Installs Docker CE from the official Docker repository, ensures the Docker service is running and enabled on boot, and adds the target user to the `docker` group.

**Variables (defaults/main.yml):**
```yaml
docker_user: vagrant
```

**Handlers (handlers/main.yml):**
- `restart docker` — Restarts the Docker service. Triggered when Docker packages are installed or updated.

**Dependencies:** Depends on `common` role being run first (curl must be available for GPG key download).

---

### app_deploy

**Purpose:** Authenticates with Docker Hub, pulls the application image, removes any existing container, runs a fresh container with the correct port mapping, and verifies the application is healthy.

**Variables (defaults/main.yml):**
```yaml
app_port: 8000
app_restart_policy: unless-stopped
app_env_vars: {}
```

**Sensitive variables (group_vars/all.yml — Vault encrypted):**
- `dockerhub_username`
- `dockerhub_password`
- `docker_image`
- `docker_image_tag`
- `app_container_name`

**Handlers (handlers/main.yml):**
- `restart app` — Restarts the application container when triggered.

**Dependencies:** Depends on `docker` role — Docker must be installed before deploying containers.

---

## 3. Idempotency Demonstration

### First Run Output
```
PLAY [Provision web servers]
TASK [Gathering Facts] ok: [localhost]
TASK [common : Update apt cache] ok: [localhost]
TASK [common : Install common packages] changed: [localhost]
TASK [docker : Install prerequisites] ok: [localhost]
TASK [docker : Create keyrings directory] ok: [localhost]
TASK [docker : Add Docker GPG key] changed: [localhost]
TASK [docker : Add Docker repository] changed: [localhost]
TASK [docker : Install Docker packages] changed: [localhost]
TASK [docker : Ensure Docker service is running and enabled] ok: [localhost]
TASK [docker : Add user to docker group] changed: [localhost]
TASK [docker : Install python3-docker] changed: [localhost]
RUNNING HANDLER [docker : restart docker] changed: [localhost]

PLAY RECAP
localhost : ok=12  changed=7  unreachable=0  failed=0
```

### Second Run Output
```
PLAY [Provision web servers]
TASK [Gathering Facts] ok: [localhost]
TASK [common : Update apt cache] ok: [localhost]
TASK [common : Install common packages] ok: [localhost]
TASK [docker : Install prerequisites] ok: [localhost]
TASK [docker : Create keyrings directory] ok: [localhost]
TASK [docker : Add Docker GPG key] ok: [localhost]
TASK [docker : Add Docker repository] ok: [localhost]
TASK [docker : Install Docker packages] ok: [localhost]
TASK [docker : Ensure Docker service is running and enabled] ok: [localhost]
TASK [docker : Add user to docker group] ok: [localhost]
TASK [docker : Install python3-docker] ok: [localhost]

PLAY RECAP
localhost : ok=11  changed=0  unreachable=0  failed=0
```

### Analysis

**First run — what changed and why:**
- `Install common packages` — packages were not yet installed
- `Add Docker GPG key` — key file did not exist
- `Add Docker repository` — repository was not configured
- `Install Docker packages` — Docker was not installed
- `Add user to docker group` — vagrant user was not in docker group
- `Install python3-docker` — Python Docker library was not installed
- `restart docker` handler — triggered because Docker packages were installed

**Second run — why nothing changed:**
Every Ansible module checks the current state before acting. `apt` checks if packages are already present. `file` checks if the directory exists. `apt_repository` checks if the repo is already configured. `user` checks group membership. Since the desired state was already achieved on the first run, no changes were needed on the second run.

**What makes these roles idempotent:**
- Using `apt: state=present` instead of running raw install commands
- Using `file: state=directory` instead of `mkdir`
- Using `apt_repository` module which checks before adding
- Using `creates:` argument on the shell task for the GPG key — skips if file already exists
- Using `service: state=started` instead of raw `systemctl start`

---

## 4. Ansible Vault Usage

### How Credentials Are Stored

Sensitive data (Docker Hub credentials, image name, ports) are stored in `group_vars/all.yml`, encrypted with Ansible Vault. The file is safe to commit to Git because it is AES-256 encrypted.

### Vault Password Management

The vault password is never stored in the repository. It is entered interactively at runtime using `--ask-vault-pass`. In a CI/CD pipeline, it would be stored as a secret environment variable and passed via `--vault-password-file`.

### Encrypted File Example

```
$ANSIBLE_VAULT;1.1;AES256
33313938643165336263383332623738323039613932393034366566663834623931343937353161
3434396331653966343466303138646234366464393065630a616662363939653539643733336638
32333339366530373137353139313561343762313562666437303966363337633366623462326366
...
```

This is what `group_vars/all.yml` looks like in the repository — unreadable without the vault password.

### Why Ansible Vault Is Necessary

Without Vault, credentials like Docker Hub tokens would be stored in plain text in the repository, exposing them to anyone with repository access. Vault allows secrets to be version-controlled safely alongside the code that uses them, without risk of credential leakage.

---

## 5. Deployment Verification

### deploy.yml Run Output
```
TASK [app_deploy : Log in to Docker Hub] changed: [localhost]
TASK [app_deploy : Pull Docker image] ok: [localhost]
TASK [app_deploy : Stop existing container] ...ignoring (no container existed)
TASK [app_deploy : Remove old container] ok: [localhost]
TASK [app_deploy : Run application container] changed: [localhost]
TASK [app_deploy : Wait for application to be ready] ok: [localhost]
TASK [app_deploy : Verify health endpoint] ok: [localhost]

PLAY RECAP
localhost : ok=8  changed=2  unreachable=0  failed=0  ignored=1
```

### Container Status (`docker ps`)
```
CONTAINER ID   IMAGE                               COMMAND           CREATED          STATUS          PORTS                    NAMES
8376a0ef5240   3llimi/devops-info-service:latest   "python app.py"   28 seconds ago   Up 27 seconds   0.0.0.0:8000->8000/tcp   devops-info-service
```

### Health Check Verification
```bash
$ curl http://localhost:8000/health
{"status":"healthy","timestamp":"2026-02-21T02:04:28.847408+00:00","uptime_seconds":25}

$ curl http://localhost:8000/
{"service":{"name":"devops-info-service","version":"1.0.0","description":"DevOps course info service","framework":"FastAPI"},
"system":{"hostname":"8376a0ef5240","platform":"Linux",...},
"runtime":{"uptime_seconds":25,...}}
```

### Handler Execution

The `restart docker` handler in the docker role was triggered during the first provisioning run when Docker packages were installed. On subsequent runs it was not triggered because no changes were made to Docker packages — demonstrating that handlers only fire when their notifying task actually changes something.

---

## 6. Key Decisions

**Why use roles instead of plain playbooks?**
Roles enforce a standard structure that makes code reusable and maintainable. Each role can be developed, tested, and shared independently. A single monolithic playbook with all tasks mixed together would be harder to read, impossible to reuse, and difficult to test in isolation.

**How do roles improve reusability?**
Each role encapsulates all logic for a single concern — the `docker` role can be dropped into any other project that needs Docker installed, without copying individual tasks. Default variables allow roles to be customized without modifying their internals.

**What makes a task idempotent?**
A task is idempotent when it checks the current state before acting and only makes changes if the desired state is not already achieved. Ansible's built-in modules (apt, service, file, user) handle this automatically — unlike raw shell commands which always execute regardless of current state.

**How do handlers improve efficiency?**
Handlers only run when notified by a task that actually made a change. Without handlers, you would restart Docker after every playbook run even if nothing changed. With handlers, Docker is only restarted when packages are actually installed or updated — avoiding unnecessary service disruptions.

**Why is Ansible Vault necessary?**
Any secret stored in plain text in a Git repository is effectively public, even in private repos. Vault encrypts secrets at rest while keeping them version-controlled alongside the infrastructure code. This allows the full Ansible project (including secrets) to be committed to Git safely.

---

## 7. Challenges

- **WSL2 disk space:** The WSL2 Alpine distro had only 136MB disk space, not enough to install Ansible. Solved by installing Ansible directly on the Vagrant VM and running it against localhost.
- **Docker login module:** `community.general.docker_login` failed in Ansible 2.10. Solved by using a `shell` task with `docker login --password-stdin` instead.
- **group_vars not loading with become:** Vault-encrypted `group_vars/all.yml` variables were not accessible when `become: yes` was set at the play level. Solved by passing variables explicitly with `-e @group_vars/all.yml` and setting `become: no` in the deploy playbook.
- **App port:** The application runs on port 8000 (FastAPI/Uvicorn), not 5000 as initially assumed. Discovered via `docker logs` and corrected in the vault variables and port mapping.

---

## 8. Bonus — Dynamic Inventory

### Approach
Since no cloud provider was available, a custom Python dynamic inventory script was created (`inventory/dynamic_inventory.py`). This demonstrates the same concepts as cloud inventory plugins — hosts are discovered at runtime rather than hardcoded.

### How It Works
The script runs at playbook execution time, queries the system for hostname and IP dynamically, and outputs a JSON inventory structure that Ansible consumes. This means if the VM's hostname or IP changes, the inventory automatically reflects the new values without any manual updates.

### ansible-inventory --graph Output
```
@all:
  |--@ungrouped:
  |--@webservers:
  |  |--localhost
```

### Running Playbooks with Dynamic Inventory
```bash
ansible all -i inventory/dynamic_inventory.py -m ping --ask-vault-pass
# localhost | SUCCESS => { "ping": "pong" }

ansible-playbook playbooks/provision.yml -i inventory/dynamic_inventory.py --ask-vault-pass
# localhost : ok=11  changed=1  unreachable=0  failed=0
```

### Benefits vs Static Inventory
With static inventory, if the VM IP or hostname changes you must manually edit `hosts.ini`. With dynamic inventory, the script queries the system at runtime so it always reflects the current state. In a cloud environment with auto-scaling, this is essential — new VMs appear and disappear constantly and maintaining a static file would be impossible.