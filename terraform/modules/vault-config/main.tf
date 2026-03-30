resource "vault_policy" "name" {
  name = "${var.service_name}-policy"
  policy = <<EOT
path "secret/data/${var.service_name}/*" {
  capabilities = ["read"]
}
%{ for path in var.extra_paths ~}
path "${path}" {
  capabilities = ["read"]
}
%{ endfor ~}
%{ if var.enable_pki ~}
path "${var.pki_backend}/issue/${var.pki_role != "" ? var.pki_role : "${var.service_name}-pki"}" {
  capabilities = ["create", "update"]
}
%{ endif ~}
EOT
}

resource "vault_aws_auth_backend_role" "name" {
  count = var.auth_type == "aws" ? 1 : 0
  backend = var.auth_backend_path
  role  = "${var.service_name}-role"
  auth_type  = "iam"
  bound_iam_principal_arns = ["arn:aws:iam::${var.aws_account_id}:role/${var.iam_role_name}"]
  token_policies = [vault_policy.name.name]
  token_ttl = var.token_ttl
  token_max_ttl = var.token_max_ttl
}

resource "vault_kubernetes_auth_backend_role" "name" {
  count = var.auth_type == "kubernetes" ? 1 : 0
  backend = var.auth_backend_path
  role_name = "${var.service_name}-role"
  bound_service_account_names = [var.k8s_service_account]
  bound_service_account_namespaces = [var.k8s_namespace]
  token_policies = [vault_policy.name.name]
  token_ttl = var.token_ttl
  token_max_ttl = var.token_max_ttl
}

resource "vault_pki_secret_backend_role" "name" {
  count            = var.enable_pki ? 1 : 0
  backend          = var.pki_backend
  name             = var.pki_role != "" ? var.pki_role : "${var.service_name}-pki"
  ttl              = var.token_ttl
  max_ttl          = var.token_max_ttl
  allow_ip_sans    = true
  allowed_domains  = var.allowed_domains
  allow_subdomains = true
}
