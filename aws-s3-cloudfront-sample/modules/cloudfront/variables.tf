variable "bucket_regional_domain_name" {
  description = "S3バケットのリージョン別ドメイン名（Origin設定用）"
  type        = string
}

variable "bucket_arn" {
  description = "S3バケットARN（バケットポリシー参照用）"
  type        = string
}

variable "environment" {
  description = "環境名（dev/stg/prod）"
  type        = string
  default     = "dev"
}

variable "price_class" {
  description = "CloudFrontの価格クラス（PriceClass_100/200/All）"
  type        = string
  default     = "PriceClass_200"
}
