terraform {
  backend "s3" {
    bucket = "transcendance-secrets-437836833311"
    key    = "terraform.tfstate"
    region = "eu-north-1"
  }
}
