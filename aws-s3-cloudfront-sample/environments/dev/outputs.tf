output "s3_bucket_id" {
  description = "S3バケットID"
  value       = module.s3.bucket_id
}

output "s3_bucket_arn" {
  description = "S3バケットARN"
  value       = module.s3.bucket_arn
}

output "s3_bucket_regional_domain_name" {
  description = "S3バケットのリージョン別ドメイン名"
  value       = module.s3.bucket_regional_domain_name
}

output "cloudfront_distribution_url" {
  description = "CloudFrontのURL"
  value       = "https://${module.cloudfront.distribution_domain_name}"
}
