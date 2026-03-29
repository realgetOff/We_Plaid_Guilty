output "policy_name" {
  value = vault_policy.name.name
}

output "role_name" {
  value = var.service_name
}
