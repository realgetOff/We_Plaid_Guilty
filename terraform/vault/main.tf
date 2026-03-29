resource "vault_auth_backend" "aws" {
  type = "aws"
}

resource "vault_auth_backend" "kubernetes" {
  type = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "config" {
  backend         = vault_auth_backend.kubernetes.path
  kubernetes_host = "https://${data.terraform_remote_state.infra.outputs.master_ip}:6443"
}

module "vault_app" {
  source              = "../modules/vault-config"
  service_name        = "app"
  auth_type           = "kubernetes"
  auth_backend_path   = vault_auth_backend.kubernetes.path
  k8s_service_account = "default"
  k8s_namespace       = "default"
  token_ttl           = 3600
  token_max_ttl       = 86400
  aws_account_id      = data.terraform_remote_state.infra.outputs.aws_account_id
  extra_paths         = ["secret/data/db/*"]
}

module "vault_db" {
  source            = "../modules/vault-config"
  service_name      = "db"
  auth_type         = "aws"
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name     = data.terraform_remote_state.infra.outputs.k3s_role_name
  aws_account_id    = data.terraform_remote_state.infra.outputs.aws_account_id
  token_ttl         = 3600
  token_max_ttl     = 86400
}
