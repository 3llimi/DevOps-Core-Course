variable "vm_host" {
  description = "VM IP address"
  default     = "127.0.0.1"
}

variable "vm_port" {
  description = "SSH port"
  default     = 2222
}

variable "vm_user" {
  description = "SSH username"
  default     = "vagrant"
}

variable "ssh_private_key_path" {
  description = "Path to SSH private key"
  default     = "../vagrant/.vagrant/machines/default/virtualbox/private_key"
}
