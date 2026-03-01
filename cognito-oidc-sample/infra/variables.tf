variable "aws_region" {
  type    = string
  default = "ap-northeast-1"
}

variable "project_name" {
  type    = string
  default = "cognito-oidc-sample"
}

variable "callback_url" {
  type    = string
  default = "http://localhost:8080/callback"
}

variable "logout_url" {
  type    = string
  default = "http://localhost:3000"
}
