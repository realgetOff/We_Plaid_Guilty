variable "aws_region" {
  description = "Region AWS"
  type = string
  default = "eu-north-1"
}

variable "project_name" {
  description = "Nom du projet"
  type = string
}

variable "admin_public_key" {
  description = "Clé SSH admin"
  type        = string
}

variable "email" {
  description = "email-alert"
  type        = string
}

variable "vault_root_token" {
  description = "Root token Vault"
  type = string
  sensitive = true //WARN INPORTANT
}
