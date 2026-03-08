terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 6.0"
    }
    local = {
      source = "hashicorp/local"
      version = "~> 2.0"
    }
    vault = {
      source = "hashicorp/vault"
      version = "~> 4.0"
    }
  }
}

provider "vault" {
  address = "http://${module.app.public_ip}:8200"
  token = var.vault_root_token
}

provider "aws" {
  region = var.aws_region
}
