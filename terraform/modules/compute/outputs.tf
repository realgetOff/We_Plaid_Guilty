output "public_ip" {
  description = "Public IP for Ansible"
  value = aws_instance.name.public_ip
}

output "instance_id" {
  description = "Instance ID for (debug/ref)"
  value = aws_instance.name.id
}

output "private_ip" {
  value = aws_instance.name.private_ip
}
