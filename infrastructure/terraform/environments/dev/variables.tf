variable "aws_region" {
  description = "AWS Region."
  type        = string
  default     = "eu-north-1"
}

variable "project_name" {
  description = "Project name."
  type        = string
  default     = "cloud-reliability"
}

variable "environment" {
  description = "Deployment environment."
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.10.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR block for the public subnet."
  type        = string
  default     = "10.10.1.0/24"
}

variable "availability_zone" {
  description = "Availability Zone."
  type        = string
  default     = "eu-north-1a"
}

variable "admin_cidr" {
  description = "Your current public IPv4 address in /32 notation."
  type        = string
}
