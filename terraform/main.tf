terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

# Generate SSH key pair for VM access
resource "null_resource" "vm_setup" {
  connection {
    type        = "ssh"
    host        = var.vm_host
    port        = var.vm_port
    user        = var.vm_user
    private_key = file(var.ssh_private_key_path)
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get update -y",
      "sudo apt-get install -y curl wget",
      "echo 'VM provisioned by Terraform' > /home/vagrant/terraform_managed.txt",
    ]
  }
}