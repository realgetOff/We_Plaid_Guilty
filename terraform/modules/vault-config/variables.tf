variable "aws_account_id" {
  description = "for ARN IAM"
  type = string
}

variable "iam_role_name" {
  description = "vault-kms-role"
  type = string
}

variable "token_ttl" {
  description = "time token"
  type = number
}

variable "token_max_ttl" {
  description = "time max token"
  type = number
}

variable "service_name" {
  description = "Nom du service lie a Vault"
  type = string
}

variable "auth_backend_path" {
  description = "auth backend path"
  type = string
}

variable "extra_paths" {
  description = "Paths supp fro policy"
  type        = list(string)
  default     = []
}
