locals {
  name_prefix = "${var.project_name}-${var.environment}"

  roles = toset([
    "application",
    "prometheus",
    "grafana"
  ])

  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# Resolve the latest Amazon Linux 2023 x86_64 AMI
data "aws_ssm_parameter" "amazon_linux_2023" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

# Register our local public key with EC2
resource "aws_key_pair" "admin" {
  key_name_prefix = "${local.name_prefix}-"
  public_key      = var.ssh_public_key

  tags = local.common_tags
}

resource "aws_instance" "service" {
  for_each = local.roles

  ami                    = data.aws_ssm_parameter.amazon_linux_2023.value
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [var.security_group_ids[each.key]]
  key_name               = aws_key_pair.admin.key_name

  associate_public_ip_address = true

  # Require IMDSv2 for instance metadata requests
  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  # Avoid surplus CPU-credit charges
  credit_specification {
    cpu_credits = "standard"
  }

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size
    encrypted             = true
    delete_on_termination = true
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-${each.key}"
    Role = each.key
  })

  volume_tags = merge(local.common_tags, {
    Role = each.key
  })
}
