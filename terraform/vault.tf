resource "vault_auth_backend" "aws" {
  type = "aws"
}

module "vault_app" {
  source = "./modules/vault-config"
  service_name = "app"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.aws_caller_identity.current.account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = aws_iam_role.vault_kms.name
}

module "elk" {
  source = "./modules/vault-config"
  service_name = "elk"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.aws_caller_identity.current.account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = aws_iam_role.vault_kms.name
}

module "grafana" {
  source = "./modules/vault-config"
  service_name = "app"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.aws_caller_identity.current.account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = aws_iam_role.vault_kms.name
}
