resource "aws_security_group" "app_sg" {
  name = "${var.project_name}-app_sg"
  description = "Security group: SSH (22), HTTP (80), Vault (8200)"
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    protocol = "tcp"
    to_port = 22
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 80
    protocol = "tcp"
    to_port = 80
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 8200
    protocol = "tcp"
    to_port = 8200
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
  description = "Security group:SSH (22), Kibana (5601), Grafana (3000), Elasticsearch (9200)"
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    protocol = "tcp"
    to_port = 22
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 5601
    protocol = "tcp"
    to_port = 5601
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 3000
    protocol = "tcp"
    to_port = 3000
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 9200
    protocol = "tcp"
    to_port = 9200
  }
  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    protocol = "-1"
    to_port = 0
  }
}
