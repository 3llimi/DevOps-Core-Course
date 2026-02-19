import pulumi
import subprocess
from pulumi_command import local

# Configuration
config = pulumi.Config()
vm_host = config.get("vm_host") or "127.0.0.1"
vm_port = config.get("vm_port") or "2222"
vm_user = config.get("vm_user") or "vagrant"
ssh_key_path = "C:/Users/3llim/OneDrive/Documents/GitHub/DevOps-Core-Course/vagrant/.vagrant/machines/default/virtualbox/private_key"

# Provision the VM using subprocess
vm_setup = local.Command("vm-setup",
    create=f'ssh -p {vm_port} -i "{ssh_key_path}" -o StrictHostKeyChecking=no {vm_user}@{vm_host} "touch /home/vagrant/pulumi_managed.txt"',
    delete=f'ssh -p {vm_port} -i "{ssh_key_path}" -o StrictHostKeyChecking=no {vm_user}@{vm_host} "rm -f /home/vagrant/pulumi_managed.txt"',
    interpreter=["powershell", "-Command"]
)

# Outputs
pulumi.export("vm_host", vm_host)
pulumi.export("vm_port", vm_port)
pulumi.export("vm_user", vm_user)
pulumi.export("connection_command", f"ssh -p {vm_port} {vm_user}@{vm_host} -i {ssh_key_path}")