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
