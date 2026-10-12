output "application_security_group_id" {
  description = "Security group for Go application."
  value       = aws_security_group.service["application"].id
}

output "prometheus_security_group_id" {
  description = "Security group for Prometheus."
  value       = aws_security_group.service["prometheus"].id
}

output "grafana_security_group_id" {
  description = "Security group for Grafana."
  value       = aws_security_group.service["grafana"].id
}

output "security_group_ids" {
  description = "All service security groups keyed by role."
  value = {
    for role, sg in aws_security_group.service : role => sg.id
  }
}
