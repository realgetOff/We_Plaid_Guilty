output "master_ip" {
  value = module.master.public_ip
}
output "worker1_ip" {
  value = module.worker1.public_ip
}

output "worker2_ip" {
  value = module.worker2.public_ip
}

output "connect_command_to_master" { //APP
  value = "ssh ec2-user@${module.master.public_ip}"
}

output "connect_command_to_worker1" { //ELK
  value = "ssh ec2-user@${module.worker1.public_ip}"
}

output "connect_command_to_worker2" { //GRAFANE
  value = "ssh ec2-user@${module.worker2.public_ip}"
}

output "kms_key_id" {
  value = aws_kms_key.vault_unseal.key_id
}

output "aws_account_id" {
  value = data.aws_caller_identity.current.account_id
}
/*
output "vault_kms_role_name" {
  value = aws_iam_role.vault_kms.name
}

output "vault_reader_elk_role_name" {
  value = aws_iam_role.vault_reader_elk.name
}*/

output "k3s_role_name" {
  value = aws_iam_role.k3s.name
}
