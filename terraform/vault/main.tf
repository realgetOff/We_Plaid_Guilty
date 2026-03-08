resource "vault_auth_backend" "aws" {
  type = "aws"
}

module "vault_app" {
  source = "./modules/vault-config"
  service_name = "app"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.terraform_remote_state.infra.outputs.aws_account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = data.terraform_remote_state.infra.outputs.vault_kms_role_name
}

module "vault_elk" {
  source = "./modules/vault-config"
  service_name = "elk"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.terraform_remote_state.infra.outputs.aws_account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = data.terraform_remote_state.infra.outputs.vault_reader_elk_role_name
}

module "vault_grafana" {
  source = "./modules/vault-config"
  service_name = "grafana"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.terraform_remote_state.infra.outputs.aws_account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = data.terraform_remote_state.infra.outputs.vault_reader_grafana_role_name
}
