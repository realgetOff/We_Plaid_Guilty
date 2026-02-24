resource "aws_s3_bucket" "secrets" {
  bucket = "transcendance-secrets-${data.aws_caller_identity.current.account_id}"
  force_destroy = true
  tags   = { Name = var.project_name }
}

resource "aws_s3_bucket_versioning" "secrets" {
  bucket = aws_s3_bucket.secrets.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "secrets" {
  bucket = aws_s3_bucket.secrets.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.vault_unseal.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "secrets" {
  bucket                  = aws_s3_bucket.secrets.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
