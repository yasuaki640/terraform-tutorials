module "vpc" {
  source = "../../modules/vpc"

  name_prefix         = var.environment
  vpc_cidr            = var.vpc_cidr
  availability_zone   = var.availability_zone
  public_subnet_cidr  = var.public_subnet_cidr
  private_subnet_cidr = var.private_subnet_cidr
}
