variable "aws_account_id" {
  description = "for ARN IAM"
  type = string
  default = ""
}

variable "iam_role_name" {
  description = "vault-kms-role"
  type = string
  default = ""
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

variable "auth_type" {
  description = "aws kubernetes ?"
  type        = string
  default     = "aws"
}

variable "k8s_service_account" {
  description = "auth_type = kubernetes"
  type        = string
  default     = "default"
}

variable "k8s_namespace" {
  description = "auth_type = kubernetes"
  type        = string
  default     = "default"
}

variable "enable_pki" {
  description = "Enable PKI cert injection service"
  type        = bool
  default     = false
}

variable "pki_backend" {
  description = "PKI backend path"
  type        = string
  default     = "pki"
}

variable "pki_role" {
  description = "PKI role name"
  type        = string
  default     = ""
}

variable "allowed_domains" {
  description = "allowed domains PKI certs"
  type        = list(string)
  default     = []
}
