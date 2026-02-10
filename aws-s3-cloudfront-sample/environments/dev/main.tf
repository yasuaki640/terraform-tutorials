module "s3" {
  source = "../../modules/s3"

  bucket_name = var.bucket_name
  environment = var.environment
}

module "cloudfront" {
  source = "../../modules/cloudfront"

  bucket_regional_domain_name = module.s3.bucket_regional_domain_name
  bucket_arn                  = module.s3.bucket_arn
}

resource "aws_s3_bucket_policy" "cloudfront_access" {
  bucket = module.s3.bucket_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontServicePrincipal"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = "s3:GetObject"
        Resource  = "${module.s3.bucket_arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = module.cloudfront.distribution_arn
          }
        }
      }
    ]
  })
}
