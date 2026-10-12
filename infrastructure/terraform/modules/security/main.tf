locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  security_groups = {
    application = "Go application and metrics"
    prometheus  = "Prometheus and Alertmanager"
    grafana     = "Grafana dashboards and host metrics"
  }
}

resource "aws_security_group" "service" {
  for_each = local.security_groups

  name_prefix = "${local.name_prefix}-${each.key}-"
  description = each.value
  vpc_id      = var.vpc_id

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-${each.key}-sg"
    Role = each.key
  })

  lifecycle {
    create_before_destroy = true
  }
}

# SSH administration from developer workstation
resource "aws_vpc_security_group_ingress_rule" "ssh" {
  for_each = aws_security_group.service

  security_group_id = each.value.id
  description       = "SSH from administrator public IP"

  cidr_ipv4   = var.admin_cidr
  ip_protocol = "tcp"
  from_port   = 22
  to_port     = 22
}

# Expose selected service UIs only to administrator
resource "aws_vpc_security_group_ingress_rule" "admin_services" {
  for_each = {
    application = 8080
    prometheus  = 9090
    grafana     = 3000
  }

  security_group_id = aws_security_group.service[each.key].id
  description       = "Service access from administrator"

  cidr_ipv4   = var.admin_cidr
  ip_protocol = "tcp"
  from_port   = each.value
  to_port     = each.value
}

# Prometheus scrapes Go application metrics on /metrics
resource "aws_vpc_security_group_ingress_rule" "prometheus_to_app" {
  security_group_id = aws_security_group.service["application"].id
  description       = "Prometheus scrapes Go API metrics"

  referenced_security_group_id = aws_security_group.service["prometheus"].id

  ip_protocol = "tcp"
  from_port   = 8080
  to_port     = 8080
}

# Prometheus scrapes Node Exporter on all three VMs
resource "aws_vpc_security_group_ingress_rule" "node_exporter" {
  for_each = aws_security_group.service

  security_group_id = each.value.id
  description       = "Prometheus scrapes Node Exporter"

  referenced_security_group_id = aws_security_group.service["prometheus"].id

  ip_protocol = "tcp"
  from_port   = 9100
  to_port     = 9100
}

# Grafana queries the Prometheus HTTP API
resource "aws_vpc_security_group_ingress_rule" "grafana_to_prometheus" {
  security_group_id = aws_security_group.service["prometheus"].id
  description       = "Grafana queries Prometheus"

  referenced_security_group_id = aws_security_group.service["grafana"].id

  ip_protocol = "tcp"
  from_port   = 9090
  to_port     = 9090
}

# Initial outbound access for software installation and updates
resource "aws_vpc_security_group_egress_rule" "outbound_ipv4" {
  for_each = aws_security_group.service

  security_group_id = each.value.id
  description       = "Outbound IPv4 access"

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}
