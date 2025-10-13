module "vpc" {
  source = "../../modules/vpc"

  name_prefix = var.environment
  vpc_cidr    = var.vpc_cidr
}