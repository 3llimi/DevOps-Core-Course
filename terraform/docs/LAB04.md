# Lab 04 — Infrastructure as Code (Terraform & Pulumi)

## 1. Cloud Provider & Infrastructure

- **Cloud Provider:** Local VM (VirtualBox + Vagrant)
- **Reason:** No cloud provider access available from Russia (free Yandex Cloud credits were used in a previous course)
- **Instance:** Ubuntu 22.04 LTS (jammy64)
- **Resources Created:**
  - Vagrant VM (2GB RAM, 38GB disk)
  - Private network (192.168.56.10)
  - SSH access via port 2222
- **Total Cost:** $0

---

## 2. Terraform Implementation

- **Terraform Version:** 1.9.8 (windows_amd64)
- **Provider:** hashicorp/null v3.2.3 + integrations/github v6.6.0

### Project Structure
```
terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── github.tf
└── .gitignore
```

### Key Decisions
- Used null provider with remote-exec provisioner since no cloud provider was available
- SSH key path points to Vagrant-generated private key
- Variables used for VM host, port, user, and SSH key path
- Provider installed manually due to registry.terraform.io being blocked in Russia
- GitHub provider added for bonus task (repository import)

### Challenges
- `registry.terraform.io` is blocked in Russia — had to download providers manually and use `-plugin-dir` flag
- Terraform was installed as 32-bit by default via winget — had to reinstall AMD64 version manually

### terraform init output
```
Initializing the backend...
Initializing provider plugins...
- Finding hashicorp/null versions matching "~> 3.0"...
- Installing hashicorp/null v3.2.3...
- Installed hashicorp/null v3.2.3 (unauthenticated)

Terraform has been successfully initialized!
```

### terraform plan output
```
github_repository.course_repo: Refreshing state... [id=DevOps-Core-Course]

Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # null_resource.vm_setup will be created
  + resource "null_resource" "vm_setup" {
      + id = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + connection_command = "ssh -p 2222 vagrant@127.0.0.1 -i ../vagrant/.vagrant/machines/default/virtualbox/private_key"
  + vm_host            = "127.0.0.1"
  + vm_port            = 2222
  + vm_user            = "vagrant"
```

### terraform apply output
```
null_resource.vm_setup: Creating...
null_resource.vm_setup: Provisioning with 'remote-exec'...
null_resource.vm_setup (remote-exec): Connecting to remote host via SSH...
null_resource.vm_setup (remote-exec):   Host: 127.0.0.1
null_resource.vm_setup (remote-exec):   User: vagrant
null_resource.vm_setup (remote-exec):   Private key: true
null_resource.vm_setup (remote-exec): Connected!
null_resource.vm_setup (remote-exec): Fetched 8922 kB in 5s (1911 kB/s)
null_resource.vm_setup (remote-exec): curl is already the newest version (7.81.0-1ubuntu1.21).
null_resource.vm_setup (remote-exec): wget is already the newest version (1.21.2-2ubuntu1.1).
null_resource.vm_setup (remote-exec): 0 upgraded, 0 newly installed, 0 to remove and 1 not upgraded.
null_resource.vm_setup: Creation complete after 32s [id=3159720517304979827]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:
connection_command = "ssh -p 2222 vagrant@127.0.0.1 -i ../vagrant/.vagrant/machines/default/virtualbox/private_key"
vm_host = "127.0.0.1"
vm_port = 2222
vm_user = "vagrant"
```

### SSH Access Proof
```
$ ssh -p 2222 vagrant@127.0.0.1 -i "../vagrant/.vagrant/machines/default/virtualbox/private_key"
Welcome to Ubuntu 22.04.5 LTS (GNU/Linux 5.15.0-170-generic x86_64)

Last login: Thu Feb 19 19:03:33 2026 from 10.0.2.2
vagrant@ubuntu-jammy:~$ cat ~/terraform_managed.txt
VM provisioned by Terraform
```
### terraform destroy output
```
null_resource.vm_setup: Destroying... [id=8395842967608656684]
null_resource.vm_setup: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.
```
---

## 3. Pulumi Implementation

- **Pulumi Version:** v3.222.0
- **Language:** Python
- **Provider:** pulumi-command v1.1.3

### Project Structure
```
pulumi/
├── __main__.py
├── requirements.txt
├── Pulumi.yaml
└── venv/
```

### Key Differences from Terraform
- Infrastructure defined in Python instead of HCL
- Used `pulumi_command.local.Command` to run SSH commands on the VM
- State stored locally using `pulumi login --local` (no Pulumi Cloud needed)
- Required `interpreter=["powershell", "-Command"]` for Windows compatibility
- Python venv needed before any code runs — extra setup step vs Terraform

### Challenges
- Import path for pulumi-command is `pulumi_command` not `pulumi.command`
- Windows SSH quoting issues — bash redirect `>` didn't work through cmd/PowerShell
- Had to use `touch` instead of `echo` to avoid shell quoting problems

### pulumi preview output
```
Previewing update (dev):
     Type                      Name              Plan
 +   pulumi:pulumi:Stack       lab04-pulumi-dev  create
 +   └─ command:local:Command  vm-setup          create

Outputs:
    connection_command: "ssh -p 2222 vagrant@127.0.0.1 -i C:/Users/.../private_key"
    vm_host           : "127.0.0.1"
    vm_port           : "2222"
    vm_user           : "vagrant"

Resources:
    + 2 to create
```

### pulumi up output
```
Updating (dev):
     Type                      Name              Status
 +   pulumi:pulumi:Stack       lab04-pulumi-dev  created (3s)
 +   └─ command:local:Command  vm-setup          created (2s)

Outputs:
    connection_command: "ssh -p 2222 vagrant@127.0.0.1 -i C:/Users/.../private_key"
    vm_host           : "127.0.0.1"
    vm_port           : "2222"
    vm_user           : "vagrant"

Resources:
    + 2 created
Duration: 5s
```

### SSH Access Proof
```
$ ssh -p 2222 vagrant@127.0.0.1 -i "C:/Users/.../private_key"
Welcome to Ubuntu 22.04.5 LTS (GNU/Linux 5.15.0-170-generic x86_64)

vagrant@ubuntu-jammy:~$ cat /home/vagrant/pulumi_managed.txt
(empty file - created by touch command via Pulumi SSH provisioner,
proving Pulumi successfully connected and provisioned the VM)
```

---

## 4. Terraform vs Pulumi Comparison

**Ease of Learning:**
Terraform was easier to learn for simple infrastructure. HCL is declarative and purpose-built for infrastructure definition, you describe *what* you want and Terraform figures out *how*. Pulumi required more upfront setup (Python venv, pip packages, Pulumi login, passphrase) before writing any infrastructure code.

**Code Readability:**
Terraform HCL is more readable for infrastructure — it clearly describes resources and their relationships. Pulumi Python feels more familiar if you already know Python, but it's more verbose for simple tasks and requires understanding both Python and Pulumi's resource model.

**Debugging:**
Pulumi was harder to debug, errors mixed Python tracebacks with Pulumi internals, and Windows shell quoting issues made SSH commands tricky. Terraform errors were more descriptive and pointed directly to the problematic resource or argument.

**Documentation:**
Terraform has better documentation, more Stack Overflow answers, and more community examples. Pulumi docs are good but harder to find practical Windows-specific examples for common use cases.

**Use Case:**
- Use **Terraform** for straightforward cloud infrastructure provisioning where declarative style is a good fit
- Use **Pulumi** when you need complex logic, dynamic resource creation, loops, or reusable functions that are hard to express in HCL

---

## 5. Bonus Tasks

### Part 1: GitHub Actions CI/CD for Terraform (1.5 pts)

Created `.github/workflows/terraform-ci.yml` that:
- Triggers **only** on changes to `terraform/**` files (path filter)
- Runs `terraform fmt -check` — validates code formatting
- Runs `terraform init -backend=false` — initializes without state backend
- Runs `terraform validate` — checks syntax and configuration
- Runs `tflint` — lints for best practices and potential errors

**Why this matters:**
Automated validation catches syntax errors, formatting issues, and bad practices before they reach the main branch. Infrastructure changes are validated the same way application code is — through CI.

### Part 2: GitHub Repository Import (1 pt)

Added GitHub provider to Terraform and imported the existing course repository:

**Provider config (`github.tf`):**
```hcl
provider "github" {
  # token auto-detected from GITHUB_TOKEN environment variable
}

resource "github_repository" "course_repo" {
  name         = "DevOps-Core-Course"
  description  = "🚀Production-grade DevOps course..."
  visibility   = "public"
  has_issues   = false
  has_wiki     = true
  has_downloads = true
  has_projects  = true
}
```

**Import command and output:**
```
$ terraform import github_repository.course_repo DevOps-Core-Course

github_repository.course_repo: Importing from ID "DevOps-Core-Course"...
github_repository.course_repo: Import prepared!
  Prepared github_repository for import
github_repository.course_repo: Refreshing state... [id=DevOps-Core-Course]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

**After import — terraform plan shows no changes:**
```
Plan: 1 to add, 0 to change, 0 to destroy.
(only null_resource.vm_setup remaining — github_repository has no changes)
```

**Why importing existing resources matters:**
In real-world DevOps, infrastructure is often created manually before IaC is adopted. The `terraform import` command brings those existing resources under Terraform management without recreating them. This enables version control for infrastructure changes, PR-based review workflows, audit trails, and consistent configuration going forward. It's the standard way to migrate from "ClickOps" to Infrastructure as Code.

---

## 6. Lab 5 Preparation & Cleanup

**VM for Lab 5:**
- ✅ Keeping the Vagrant VM for Lab 5 (Ansible)
- VM accessible at `127.0.0.1:2222` via SSH
- Username: `vagrant`
- Key: `.vagrant/machines/default/virtualbox/private_key`

**Cleanup Status:**
- Terraform resources destroyed (`terraform destroy`) ✅
- Pulumi resources destroyed (`pulumi destroy`) ✅
- Vagrant VM kept running for Lab 5 ✅
- No secrets committed to Git ✅
- `.gitignore` configured correctly ✅

### Note on Local VM Limitations
Since a cloud provider was unavailable, the following cloud-specific 
resources were not provisioned but are understood conceptually:
- VPC/Network (not needed for local VM)
- Security Groups with ports 22, 80, 5000 (handled by Vagrant NAT/port forwarding)
- Public IP (VM accessible via 127.0.0.1:2222 through port forwarding)