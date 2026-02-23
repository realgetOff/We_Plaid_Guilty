
output "server_ip" {
  value = aws_instance.my_alma_server.public_ip
}

output "connect_command" {
  value = "ssh ec2-user@${aws_instance.my_alma_server.public_ip}"
}

output "kms_key_id" {
  value = aws_kms_key.vault_unseal.key_id
}

output "aws_account_id" {
  value = data.aws_caller_identity.current.account_id
}
