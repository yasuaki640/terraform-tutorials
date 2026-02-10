variable "bucket_name" {
  description = "S3バケット名"
  type        = string
}

variable "environment" {
  description = "環境名（dev/stg/prod）"
  type        = string
  default     = "dev"
}
