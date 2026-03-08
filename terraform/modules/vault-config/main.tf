resource "vault_policy" "name" {
  name = "${var.service_name}-policy"
  policy = <<-EOT
  path "secret/data/${var.service_name}/*" {
     capabilities = ["read"]
   }
   EOT
}

resource "vault_aws_auth_backend_role" "name" {
  backend = var.auth_backend_path
  role = "${var.service_name}-role"
  auth_type = "iam"
  bound_iam_principal_arns = ["arn:aws:iam::${var.aws_account_id}:role/${var.iam_role_name}"]
  token_policies = [vault_policy.name.name]
  token_ttl = var.token_ttl
  token_max_ttl = var.token_max_ttl
}
