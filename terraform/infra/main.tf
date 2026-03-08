data "aws_ami" "base_trans" {
  most_recent = true
  owners      = ["self"]
  filter {
    name   = "name"
    values = ["base-trans-*"]
  }
}

data "aws_caller_identity" "current" {}


resource "aws_key_pair" "admin_key" {
  key_name   = "formation-key"
  public_key = var.admin_public_key
}

module "app" {
  source = "../modules/compute"
  project_name = var.project_name
  instance_name = "EC2-app"
  instance_type = "t4g.medium"
  ami_id = data.aws_ami.base_trans.id
  volume_size = 8
  volume_type = "gp3"
  key_name = aws_key_pair.admin_key.key_name
  sg_list = [aws_security_group.app_sg.id]
  iam_profile = aws_iam_instance_profile.vault_kms.name
}

module "elk" {
  source = "../modules/compute"
  project_name = var.project_name
  instance_name = "EC2-elk"
  instance_type = "t4g.medium"
  ami_id = data.aws_ami.base_trans.id
  volume_size = 20
  volume_type = "gp3"
  key_name = aws_key_pair.admin_key.key_name
  sg_list = [aws_security_group.monitoring_sg.id]
  iam_profile = aws_iam_instance_profile.vault_elk.name
}

module "grafana" {
  source = "../modules/compute"
  project_name = var.project_name
  instance_name = "EC2-grafana"
  instance_type = "t4g.small"
  ami_id = data.aws_ami.base_trans.id
  volume_size = 8
  volume_type = "gp3"
  key_name = aws_key_pair.admin_key.key_name
  sg_list = [aws_security_group.monitoring_sg.id]
  iam_profile = aws_iam_instance_profile.vault_grafana.name
}

resource "local_file" "ansible_inventory" {
  filename = "${path.module}/inventory.ini"
  content  = <<-EOT
  [APP]
  ${module.app.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  [ELK]
  ${module.elk.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  [GRAFANA]
  ${module.grafana.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  EOT
}
