resource "aws_iam_role" "vault_kms" {
  name = "vault-kms-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "vault_kms" {
  role = aws_iam_role.vault_kms.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["kms:Encrypt", "kms:Decrypt", "kms:DescribeKey"]
        Resource = aws_kms_key.vault_unseal.arn
      },
      {
        Effect   = "Allow"
        Action   = ["iam:GetRole", "iam:GetInstanceProfile"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "vault_kms" {
  name = "vault-kms-profile"
  role = aws_iam_role.vault_kms.name
}

resource "aws_iam_role_policy_attachment" "ecr_read_only" {
  role = "vault-kms-role"
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  depends_on = [ aws_iam_role.vault_kms ]
}
