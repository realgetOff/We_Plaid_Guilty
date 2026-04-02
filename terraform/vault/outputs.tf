output "vault_app_policy" {
  value = module.vault_app.policy_name
}

output "vault_db_policy" {
  value = module.vault_db.policy_name
}

output "kubernetes_auth_path" {
  value = vault_auth_backend.kubernetes.path
}

output "aws_auth_path" {
  value = vault_auth_backend.aws.path
}
