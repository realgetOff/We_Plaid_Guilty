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

module "master" { //APP
  source        = "../modules/compute"
  project_name  = var.project_name
  instance_name = "EC2-master"
  instance_type = "t4g.medium"
  ami_id        = data.aws_ami.base_trans.id
  volume_size   = 8
  volume_type   = "gp3"
  key_name      = aws_key_pair.admin_key.key_name
  sg_list       = [aws_security_group.master_sg.id]
  iam_profile   = aws_iam_instance_profile.k3s.name
}

module "worker1" { //ELK
  source        = "../modules/compute"
  project_name  = var.project_name
  instance_name = "EC2-worker1"
  instance_type = "t4g.medium"
  ami_id        = data.aws_ami.base_trans.id
  volume_size   = 20
  volume_type   = "gp3"
  key_name      = aws_key_pair.admin_key.key_name
  sg_list       = [aws_security_group.worker_sg.id]
  iam_profile   = aws_iam_instance_profile.k3s.name
}

module "worker2" { //GRAFANA
  source        = "../modules/compute"
  project_name  = var.project_name
  instance_name = "EC2-worker2"
  instance_type = "t4g.small"
  ami_id        = data.aws_ami.base_trans.id
  volume_size   = 8
  volume_type   = "gp3"
  key_name      = aws_key_pair.admin_key.key_name
  sg_list       = [aws_security_group.worker_sg.id]
  iam_profile   = aws_iam_instance_profile.k3s.name
}

resource "local_file" "ansible_inventory" {
  filename = "${path.module}/inventory.ini"
  content  = <<-EOT
  [MASTER]
  ${module.master.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  [WORKERS]
  ${module.worker1.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  ${module.worker2.public_ip} ansible_user=ec2-user ansible_ssh_private_key_file=~/.ssh/github_actions
  EOT
}
