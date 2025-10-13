# variable "name_prefix" {
#   description = "Prefix for resource names"
#   type        = string
# }

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
}

# variable "public_subnet_cidr" {
#   description = "CIDR block for public subnet"
#   type        = string
# }
#
# variable "availability_zone" {
#   description = "Availability zone for subnet"
#   type        = string
# }