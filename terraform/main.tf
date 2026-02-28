data "aws_ami" "alma_9" {
  most_recent = true
  owners      = ["679593333241"]
  filter {
    name   = "name"
    values = ["AlmaLinux OS 9*aarch64*"]
  }
  filter {
    name = "architecture"
    values = ["arm64"]
  }
}

data "aws_caller_identity" "current" {}

resource "aws_security_group" "ssh_access" {
  name        = "${var.project_name}-sg"
  description = "allow ssh"
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 8200
    to_port     = 8200
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "admin_key" {
  key_name   = "formation-key"
  public_key = var.admin_public_key
}

resource "aws_instance" "my_alma_server" {
  ami                    = data.aws_ami.alma_9.id
  instance_type          = var.instance_type
  key_name               = aws_key_pair.admin_key.key_name
  vpc_security_group_ids = [aws_security_group.ssh_access.id]
  iam_instance_profile   = aws_iam_instance_profile.vault_kms.name
  tags                   = { Name = var.project_name }
}

resource "local_file" "ansible_inventory" {
  filename = "inventory.ini"
  content  = <<-EOT
    [alma]
    ${aws_instance.my_alma_server.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  EOT
}
