output "distribution_id" {
  description = "CloudFront Distribution ID（キャッシュInvalidation用）"
  value       = aws_cloudfront_distribution.this.id
}

output "distribution_arn" {
  description = "CloudFront Distribution ARN（S3バケットポリシーのCondition用）"
  value       = aws_cloudfront_distribution.this.arn
}

output "distribution_domain_name" {
  description = "CloudFrontのドメイン名（ユーザーアクセス用）"
  value       = aws_cloudfront_distribution.this.domain_name
}
