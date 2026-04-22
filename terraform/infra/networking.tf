resource "aws_security_group" "master_sg" {
  name        = "${var.project_name}-master_sg"
  description = "Allow ports on master node"
  #checkov:skip=CKV_AWS_382:K3s nodes require egress to VPC for cluster communication
  #checkov:skip=CKV_AWS_24:SSH open to 0.0.0.0/0 required for 42 school dynamic IPs override in prod

  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
  }
  ingress {
    description = "K3s API"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
  } //make kubctl
  ingress {
    description = "Vault"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 8200
    to_port     = 8200
    protocol    = "tcp"
  } //make vault
  ingress {
    description = "HTTPS ingress nginx"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30443
    to_port     = 30443
    protocol    = "tcp"
  }
  ingress {
    description = "HTTP ingress nginx"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30080
    to_port     = 30080
    protocol    = "tcp"
  }
  ingress {
    description = "etcd"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 2379
    to_port     = 2380
    protocol    = "tcp"
  }
  ingress {
    description = "Kubelet"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
  }
  ingress {
    description = "Node exporter"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
  }
  ingress {
    description = "Flannel VXLAN"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 8472
    to_port     = 8472
    protocol    = "udp"
  }
  ingress {
    description = "WireGuard"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 51820
    to_port     = 51820
    protocol    = "udp"
  }

  egress {
    description = "HTTPS outbound"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
  }
  egress {
    description = "DNS outbound"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 53
    to_port     = 53
    protocol    = "udp"
  }
  egress {
    description = "VPC interne"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
  }
}

resource "aws_security_group" "worker_sg" {
  name        = "${var.project_name}-worker_sg"
  description = "Allow ports on worker nodes"
  #checkov:skip=CKV_AWS_382:K3s nodes require egress to VPC for cluster communication
  #checkov:skip=CKV_AWS_24:SSH open to 0.0.0.0/0 required for 42 school dynamic IPs

  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
  }
  ingress {
    description = "Grafana NodePort"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30300
    to_port     = 30300
    protocol    = "tcp"
  }
  ingress {
    description = "Kibana NodePort"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30601
    to_port     = 30601
    protocol    = "tcp"
  }
  ingress {
    description = "Kubelet"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
  }
  ingress {
    description = "Node exporter"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
  }
  ingress {
    description = "Flannel VXLAN"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 8472
    to_port     = 8472
    protocol    = "udp"
  }
  ingress {
    description = "WireGuard"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 51820
    to_port     = 51820
    protocol    = "udp"
  }

  egress {
    description = "HTTPS outbound"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
  }
  egress {
    description = "DNS outbound"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 53
    to_port     = 53
    protocol    = "udp"
  }
  egress {
    description = "VPC interne"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
  }
}
