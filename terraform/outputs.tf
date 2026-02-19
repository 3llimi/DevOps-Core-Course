output "vm_host" {
  description = "VM host address"
  value       = var.vm_host
}

output "vm_port" {
  description = "VM SSH port"
  value       = var.vm_port
}

output "vm_user" {
  description = "VM SSH user"
  value       = var.vm_user
}

output "connection_command" {
  description = "Command to SSH into VM"
  value       = "ssh -p ${var.vm_port} ${var.vm_user}@${var.vm_host} -i ${var.ssh_private_key_path}"
}