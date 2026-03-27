resource "vault_auth_backend" "aws" {
  type = "aws"
}

module "vault_app" {
  source = "../modules/vault-config"
  service_name = "app"
  token_ttl = 3600
  token_max_ttl = 86400
  aws_account_id = data.terraform_remote_state.infra.outputs.aws_account_id
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name = data.terraform_remote_state.infra.outputs.k3s_role_name
  extra_paths = ["secret/data/db/*"]
}
