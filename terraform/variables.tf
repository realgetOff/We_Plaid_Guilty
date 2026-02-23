variable "aws_region" {
  description = "Region AWS"
  type = string
  default = "eu-north-1"
}

variable "instance_type" {
  type = string
  default = "t3.medium"
}

variable "project_name" {
  type = string
  default = "AlmaLinux-Trans-42"
}
