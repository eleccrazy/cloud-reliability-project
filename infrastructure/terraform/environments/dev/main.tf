module "network" {
  source = "../../modules/network"

  project_name       = var.project_name
  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  public_subnet_cidr = var.public_subnet_cidr
  availability_zone  = var.availability_zone
}

module "security" {
  source = "../../modules/security"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.network.vpc_id
  admin_cidr   = var.admin_cidr
}

module "compute" {
  source = "../../modules/compute"

  project_name       = var.project_name
  environment        = var.environment
  subnet_id          = module.network.public_subnet_id
  security_group_ids = module.security.security_group_ids

  ssh_public_key = file(pathexpand(var.ssh_public_key_path))

  instance_type    = "t3.micro"
  root_volume_size = 8
}
