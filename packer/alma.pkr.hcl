packer {
  required_plugins {
    amazon = { 
      source = "github.com/hashicorp/amazon"
      version = ">= 1.0.0"
    }
    ansible = {
      source = "github.com/hashicorp/ansible"
      version = ">= 1.0.0"
    }
  }
}

source "amazon-ebs" "alma" {
  region = "eu-north-1"
  instance_type = "t4g.small"
  ami_name = "base-trans-{{timestamp}}"
  ssh_username = "ec2-user"

  source_ami_filter {
    owners = ["679593333241"]
    most_recent = true
      filters = {
      name                = "AlmaLinux OS 9*aarch64*"
      architecture        = "arm64"
      virtualization-type = "hvm"
    }
  }
}

build {
  sources = ["source.amazon-ebs.alma"]
  provisioner "ansible" {
    playbook_file = "../ansible/packer.yml"
    roles_path = "../ansible/roles"
  }
}
