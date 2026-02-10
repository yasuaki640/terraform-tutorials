variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "bucket_name" {
  description = "S3バケット名"
  type        = string
}

variable "environment" {
  description = "環境名"
  type        = string
  default     = "dev"
}
