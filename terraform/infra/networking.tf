resource "aws_security_group" "master_sg" {
  name        = "${var.project_name}-master_sg"
  description = "Allow ports on master node"
  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
  }
  ingress {
    description = "HTTPS"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 443
    protocol    = "tcp"
    to_port     = 443
  }
  ingress {
    description = "HTTP"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 80
    protocol    = "tcp"
    to_port     = 80
  }
  ingress {
    description = "API K3s"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 6443
    protocol    = "tcp"
    to_port     = 6443
  }
  ingress {
    description = "Kubelet"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 10250
    protocol    = "tcp"
    to_port     = 10250
  }
  ingress {
    description = "etcd"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 2379
    protocol    = "tcp"
    to_port     = 2380
  }
  ingress {
    description = "Flannel"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 8472
    protocol    = "udp"
    to_port     = 8472
  }
  ingress {
    description = "WireGuard" //WARN Cilium ?
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 51820
    protocol    = "udp"
    to_port     = 51820
  }
  ingress {
    description = "Node Port"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30000
    protocol    = "tcp"
    to_port     = 32767
  }
  ingress {
    description = "HTTP ingress"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30080
    protocol    = "tcp"
    to_port     = 30080
  }
  ingress {
    description = "HTTPS ingress"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30443
    protocol    = "tcp"
    to_port     = 30443
  }
  ingress {
    description = "Vault"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 8200
    protocol    = "tcp"
    to_port     = 8200
  }
  ingress {
    description = "DB"
    cidr_blocks = ["10.42.0.0/24"]
    from_port   = 5432
    protocol    = "tcp"
    to_port     = 5432
  }
  ingress {
    description = "Node grafana"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30300
    protocol    = "tcp"
    to_port     = 30300
  } //DELETE -> NGINX
  ingress {
    description = "Node exporter"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 9100
    protocol    = "tcp"
    to_port     = 9100
  }
  egress {
    description = "Allow all outbound traffic"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_security_group" "worker_sg" {
  name        = "${var.project_name}-worker_sg"
  description = "Allow ports on ports on workers"
  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
  }
  ingress {
    description = "Node grafana"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30300
    protocol    = "tcp"
    to_port     = 30300
  } //DELETE -> NGINX
  ingress {
    description = "Kubelet"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 10250
    protocol    = "tcp"
    to_port     = 10250
  }
  ingress {
    description = "Flannel"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 8472
    protocol    = "udp"
    to_port     = 8472
  }
  ingress {
    description = "WireGuard" //WARN Cilium ?
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 51820
    protocol    = "udp"
    to_port     = 51820
  }
  ingress {
    description = "Node Port"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 30000
    protocol    = "tcp"
    to_port     = 32767
  }
  /*
  ingress {
    description = "Kibana"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 5601
    protocol = "tcp"
    to_port = 5601
  }
  ingress {
    description = "Grafana"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 3000
    protocol = "tcp"
    to_port = 3000
  }
  ingress {
    description = "Elasticsearch"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 9200
    protocol = "tcp"
    to_port = 9200
  }
  ingress {
    description = "Logstasch"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 5044
    protocol = "tcp"
    to_port = 5044
  }
  ingress {
    description = "Prometheus"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 9090
    protocol = "tcp"
    to_port = 9090
  }*/
  ingress {
    description = "Node exporter"
    cidr_blocks = ["172.31.0.0/16"]
    from_port   = 9100
    protocol    = "tcp"
    to_port     = 9100
  }
  egress {
    description = "Allow all outbound traffic"
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}
