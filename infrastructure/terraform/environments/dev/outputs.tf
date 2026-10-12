output "vpc_id" {
  value = module.network.vpc_id
}

output "public_subnet_id" {
  value = module.network.public_subnet_id
}

output "public_route_table_id" {
  value = module.network.public_route_table_id
}

output "internet_gateway_id" {
  value = module.network.internet_gateway_id
}

output "security_group_ids" {
  description = "Security group IDs indexed by application role."
  value       = module.security.security_group_ids
}
