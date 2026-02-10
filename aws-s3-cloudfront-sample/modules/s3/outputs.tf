output "bucket_id" {
  description = "S3バケットID"
  value       = aws_s3_bucket.this.id
}

output "bucket_arn" {
  description = "S3バケットARN"
  value       = aws_s3_bucket.this.arn
}

output "bucket_regional_domain_name" {
  description = "S3バケットのリージョン別ドメイン名"
  value       = aws_s3_bucket.this.bucket_regional_domain_name
}
