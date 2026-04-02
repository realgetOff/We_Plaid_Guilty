terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.0"
    }
  }
}

data "terraform_remote_state" "infra" {
  backend = "s3"
  config = {
    bucket = "transcendance-secrets-43783683331"
    key    = "terraform-infra.tfstate"
    region = "eu-north-1"
  }
}

provider "vault" {
  address = "http://${data.terraform_remote_state.infra.outputs.master_ip}:30820"
  token   = var.vault_root_token
}
