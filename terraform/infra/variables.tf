variable "aws_region" {
  description = "Region AWS"
  type        = string
  default     = "eu-north-1"
}

variable "project_name" {
  description = "Nom du projet"
  type        = string
}

variable "admin_public_key" {
  description = "Clé SSH admin"
  type        = string
}

variable "email" {
  description = "email-alert"
  type        = string
}

variable "aws_account_id" {
  description = "id aws cc"
  type        = string
}
