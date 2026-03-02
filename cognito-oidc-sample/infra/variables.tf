variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "project_name" {
  description = "プロジェクト名（Cognitoドメインプレフィックスとしても使用。AWS全体で一意である必要がある）"
  type        = string
  default     = "cognito-oidc-sample"
}

variable "callback_url" {
  description = "Cognito認証後のコールバックURL"
  type        = string
  default     = "http://localhost:8080/callback"
}

variable "logout_url" {
  description = "ログアウト後のリダイレクトURL"
  type        = string
  default     = "http://localhost:8080"
}
