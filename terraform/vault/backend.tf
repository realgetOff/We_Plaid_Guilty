# 
# terraform {
# backend "local" {}
# }
# 
terraform {
  backend "s3" {
    bucket = "transcendance-secrets-43783683331"
    key    = "terraform-vault.tfstate"
    region = "eu-north-1"
  }
}
