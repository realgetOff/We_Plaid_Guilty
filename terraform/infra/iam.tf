resource "aws_iam_role" "k3s" {
  name = "k3s-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "k3s_kms" {
  #checkov:skip=CKV_AWS_355:sts:GetCallerIdentity and ecr:GetAuthorizationToken require wildcard resource by AWS design
  role = aws_iam_role.k3s.id
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
        Resource = [
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.project_name}-*",
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:instance-profile/${var.project_name}-*"
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["sts:GetCallerIdentity"]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = ["ecr:GetAuthorizationToken"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecr_read_only" {
  role       = aws_iam_role.k3s.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  depends_on = [aws_iam_role.k3s]
}

resource "aws_iam_instance_profile" "k3s" {
  name = "k3s-profile"
  role = aws_iam_role.k3s.name
}
