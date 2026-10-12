variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "subnet_id" {
  description = "Subnet used by the EC2 instances."
  type        = string
}

variable "security_group_ids" {
  description = "Security group IDs keyed by service role."
  type        = map(string)
}

variable "ssh_public_key" {
  description = "SSH public key content registered in EC2."
  type        = string
}

variable "instance_type" {
  description = "Instance type for all three VMs."
  type        = string
  default     = "t3.micro"
}

variable "root_volume_size" {
  description = "Root EBS volume size in GiB."
  type        = number
  default     = 8
}
