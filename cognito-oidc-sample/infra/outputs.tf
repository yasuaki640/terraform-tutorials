output "user_pool_id" {
  value       = aws_cognito_user_pool.main.id
  description = "Cognito User Pool ID"
}

output "client_id" {
  value       = aws_cognito_user_pool_client.main.id
  description = "Cognito App Client ID"
}

output "cognito_domain" {
  value       = "https://${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
  description = "Cognito Hosted UI domain URL"
}

output "jwks_uri" {
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}/.well-known/jwks.json"
  description = "JWKS endpoint for JWT signature verification"
}

output "issuer" {
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}"
  description = "Token issuer URL (used for iss claim verification)"
}
