resource "aws_kms_key" "vault_unseal" {
  description = "Vault key"
  deletion_window_in_days = 10
  enable_key_rotation = true
  tags = { Name = var.project_name }
}

resource "aws_kms_alias" "vault_unseal" {
  name = "alias/vault-unseal"
  target_key_id = aws_kms_key.vault_unseal.key_id
}
