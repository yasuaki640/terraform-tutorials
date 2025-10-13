# variable "environment" {
#   description = "Environment name"
#   type        = string
#   default     = "dev"
# }
#
# variable "public_subnet_cidr" {
#   description = "CIDR block for public subnet"
#   type        = string
#   default     = "10.0.1.0/24"
# }
#
# variable "availability_zone" {
#   description = "Availability zone for subnet"
#   type        = string
#   default     = "ap-northeast-1a"
# }

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}
