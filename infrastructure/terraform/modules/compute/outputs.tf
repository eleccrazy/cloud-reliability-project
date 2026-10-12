output "instance_ids" {
  description = "EC2 instance IDs by role."
  value = {
    for role, instance in aws_instance.service :
    role => instance.id
  }
}

output "public_ips" {
  description = "Public IP addresses by role."
  value = {
    for role, instance in aws_instance.service :
    role => instance.public_ip
  }
}

output "private_ips" {
  description = "Private IP addresses by role."
  value = {
    for role, instance in aws_instance.service :
    role => instance.private_ip
  }
}

output "key_pair_name" {
  description = "AWS EC2 key pair name."
  value       = aws_key_pair.admin.key_name
}
