resource "aws_security_group" "app_sg" {
  name = "${var.project_name}-app_sg"
  description = "Allow ports on app instance"
  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    protocol = "tcp"
    to_port = 22
  }
  ingress {
    description = "HTTP"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 80
    protocol = "tcp"
    to_port = 80
  }
  ingress {
    description = "Vault"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 8200
    protocol = "tcp"
    to_port = 8200
  }
  ingress {
    description = "Node exporter"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 9100
    protocol = "tcp"
    to_port = 9100
  }
  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    protocol = "-1"
    to_port = 0
  }
}

resource "aws_security_group" "monitoring_sg" {
  name = "${var.project_name}-monitoring_sg"
  description = "Allow ports on monitoring instances"
  ingress {
    description = "SSH"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    protocol = "tcp"
    to_port = 22
  }
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
  }
  ingress {
    description = "Node exporter"
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 9100
    protocol = "tcp"
    to_port = 9100
  }
  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    protocol = "-1"
    to_port = 0
  }
}
