output "app_ip" {
  value = module.app.public_ip
}
output "elk_ip" {
  value = module.elk.public_ip
}

output "grafana_ip" {
  value = module.grafana.public_ip
}

output "connect_command_to_app" {
  value = "ssh ec2-user@${module.app.public_ip}"
}

output "connect_command_to_elk" {
  value = "ssh ec2-user@${module.elk.public_ip}"
}

output "connect_command_to_grafana" {
  value = "ssh ec2-user@${module.grafana.public_ip}"
}

output "kms_key_id" {
  value = aws_kms_key.vault_unseal.key_id
}

output "aws_account_id" {
  value = data.aws_caller_identity.current.account_id
}

output "vault_kms_role_name" {
  value = aws_iam_role.vault_kms.name
}

output "vault_reader_elk_role_name" {
  value = aws_iam_role.vault_reader_elk.name
}

output "vault_reader_grafana_role_name" {
  value = aws_iam_role.vault_reader_grafana.name
}
