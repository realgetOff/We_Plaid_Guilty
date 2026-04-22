resource "aws_instance" "name" {
  tags = {
    Name = var.instance_name
    Project = var.project_name
  }
  monitoring = true
  instance_type = var.instance_type
  ami = var.ami_id
  key_name = var.key_name
  vpc_security_group_ids = var.sg_list
  iam_instance_profile = var.iam_profile

  root_block_device {
    volume_size = var.volume_size
    volume_type = var.volume_type
  }
}
