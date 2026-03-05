variable "aws_region" {
  description = "Region AWS"
  type = string
  default = "eu-north-1"
}

variable "instance_type" {
  type = string
  default = "t4g.medium"
}

variable "project_name" {
  type = string
  default = "AlmaLinux-Trans-42"
}

variable "admin_public_key" {
  description = "Clé SSH admin"
  type        = string
}

variable "email" {
  description = "email-alert"
  type        = string
}
