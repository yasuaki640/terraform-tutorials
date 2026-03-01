# =============================================================================
# Terraform Outputs
# =============================================================================
# これらの値はアプリケーションの環境変数として使用する。
# `make env` コマンドで自動的に .env ファイルに書き出される。
# =============================================================================

output "user_pool_id" {
  description = "Cognito User Pool ID"
  value       = aws_cognito_user_pool.main.id
}

output "client_id" {
  description = "Cognito App Client ID（シークレットなし・パブリッククライアント）"
  value       = aws_cognito_user_pool_client.main.id
}

output "cognito_domain" {
  description = "Cognito Hosted UI のドメイン（認可・トークンエンドポイントのベースURL）"
  value       = "https://${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
}

output "jwks_uri" {
  description = "JWKS（JSON Web Key Set）エンドポイント - トークン署名検証用の公開鍵を取得するURL"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}/.well-known/jwks.json"
}

output "issuer" {
  description = "トークンの発行者（issuer）- JWT の iss クレームと照合して正当性を検証する"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}"
}
