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


/*
resource "aws_iam_role" "vault_reader_elk" {
  name = "vault_reader_elk_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role" "vault_reader_grafana" {
  name = "vault_reader_grafana_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}
*/
resource "aws_iam_role_policy" "k3s_kms" {
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
        Resource = "*"
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
/*
resource "aws_iam_role_policy" "vault_elk" {
  role = aws_iam_role.vault_reader_elk.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["sts:GetCallerIdentity"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy" "vault_grafana" {
  role = aws_iam_role.vault_reader_grafana.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["sts:GetCallerIdentity"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy" "ecr_auth" {
  role = aws_iam_role.vault_kms.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["ecr:GetAuthorizationToken"]
      Resource = "*"
    }]
  })
}

resource "aws_iam_instance_profile" "vault_elk" {
  name = "vault-elk-profile"
  role = aws_iam_role.vault_reader_elk.name
}

resource "aws_iam_instance_profile" "vault_grafana" {
  name = "vault-grafana-profile"
  role = aws_iam_role.vault_reader_grafana.name
}

resource "aws_iam_instance_profile" "vault_kms" {
  name = "vault-kms-profile"
  role = aws_iam_role.vault_kms.name
}
*/


resource "aws_iam_role_policy_attachment" "ecr_read_only" {
  role       = aws_iam_role.k3s.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  depends_on = [aws_iam_role.k3s]
}

resource "aws_iam_instance_profile" "k3s" {
  name = "k3s-profile"
  role = aws_iam_role.k3s.name
}
