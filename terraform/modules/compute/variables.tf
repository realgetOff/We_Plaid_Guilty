variable "instance_name" {
  description = "Instance Name"
  type = string
}

variable "instance_type" {
  description = "Type of instance"
  type = string
}

variable "ami_id" {
  description = "AMI ID"
  type = string
}

variable "volume_size" {
  description = "Disk Size"
  type = number
}

variable "key_name" {
  description = "AWS key name"
  type = string
}

variable "project_name" {
  description = "Project Name"
  type = string
}

variable "sg_list" {
  description = "Security Group IDS"
  type = list(string)
}

variable "iam_profile" {
  description = "IAM instance profil"
  type = string
}

variable "volume_type" {
  description = "Type of volume"
  default = "gp3"
  type = string
}
