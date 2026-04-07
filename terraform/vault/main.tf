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

resource "vault_mount" "pki" {
  path                      = "pki"
  type                      = "pki"
  default_lease_ttl_seconds = 86400
  max_lease_ttl_seconds     = 315360000
}

resource "vault_pki_secret_backend_root_cert" "root" {
  backend     = vault_mount.pki.path
  type        = "internal"
  common_name = "trans-ca"
  ttl         = "315360000"
  issuer_name = "trans-root"
}

resource "vault_pki_secret_backend_config_urls" "config" {
  backend                 = vault_mount.pki.path
  issuing_certificates    = ["http://vault.vault.svc.cluster.local:8200/v1/pki/ca"]
  crl_distribution_points = ["http://vault.vault.svc.cluster.local:8200/v1/pki/crl"]
}

module "vault_app" {
  source              = "../modules/vault-config"
  service_name        = "app"
  auth_type           = "kubernetes"
  auth_backend_path   = vault_auth_backend.kubernetes.path
  k8s_service_account = "gartic-app"
  k8s_namespace       = "app"
  token_ttl           = 40000
  token_max_ttl       = 86400
  extra_paths         = ["secret/data/db/*"]
  enable_pki          = true
  pki_backend         = vault_mount.pki.path
  allowed_domains     = ["transcendance.local", "svc.cluster.local", "app.svc.cluster.local", "default.svc.cluster.local"]
}

module "vault_db" {
  source            = "../modules/vault-config"
  service_name      = "db"
  auth_type         = "aws"
  auth_backend_path = vault_auth_backend.aws.path
  iam_role_name     = data.terraform_remote_state.infra.outputs.k3s_role_name
  aws_account_id    = data.terraform_remote_state.infra.outputs.aws_account_id
  token_ttl         = 40000
  token_max_ttl     = 86400
  enable_pki        = true
  pki_backend         = vault_mount.pki.path
  allowed_domains = ["transcendance.local", "postgres.transcendance.local"]
}

module "vault_nginx" {
  source              = "../modules/vault-config"
  service_name        = "nginx"
  auth_type           = "kubernetes"
  auth_backend_path   = vault_auth_backend.kubernetes.path
  k8s_service_account = "ingress-nginx"
  k8s_namespace       = "ingress-nginx"
  token_ttl           = 40000
  token_max_ttl       = 86400
  enable_pki          = true
  pki_backend         = vault_mount.pki.path
  allowed_domains     = ["transcendance.local", "ingress-nginx.svc.cluster.local"]
}
