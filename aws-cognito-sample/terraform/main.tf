terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
}

variable "google_client_id" {
  description = "Google OAuth Client ID"
  type        = string
  sensitive   = true
}

variable "google_client_secret" {
  description = "Google OAuth Client Secret"
  type        = string
  sensitive   = true
}

provider "aws" {
  region = "ap-northeast-1"
}

resource "aws_cognito_user_pool" "pool" {
  name = "aws-cognito-sample"

  auto_verified_attributes = ["email"]
  deletion_protection      = "ACTIVE"
  mfa_configuration        = "OPTIONAL"
  tags_all                 = {}
  username_attributes      = ["email"]

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  email_configuration {
    configuration_set     = null
    email_sending_account = "COGNITO_DEFAULT"
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = true
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      max_length = "2048"
      min_length = "0"
    }
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = true
    name                     = "customopt"
    required                 = false

    string_attribute_constraints {
      max_length = null
      min_length = "1"
    }
  }

  username_configuration {
    case_sensitive = false
  }

  software_token_mfa_configuration {
    enabled = true
  }

  user_attribute_update_settings {
    attributes_require_verification_before_update = [
      "email",
    ]
  }

  password_policy {
    minimum_length                   = 8
    require_lowercase                = true
    require_numbers                  = true
    require_symbols                  = false
    require_uppercase                = true
    temporary_password_validity_days = 7
  }
}

resource "aws_cognito_user_pool_client" "client" {
  name         = "aws-cognito-sample-application-client"
  user_pool_id = aws_cognito_user_pool.pool.id

  # nonceを有効にするために必要な設定
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["email", "openid", "profile"]

  # OpenID Connect設定を追加
  enable_token_revocation       = true
  prevent_user_existence_errors = "ENABLED"

  # OIDCのnonce対応を有効化
  auth_session_validity = 3  # minutes
  id_token_validity     = 60 # minutes

  # TODO: tfvars で指定する
  callback_urls        = ["http://localhost:8080/callback"]
  logout_urls          = ["http://localhost:8080/logout"]
  default_redirect_uri = "http://localhost:8080/callback"

  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_PASSWORD_AUTH"
  ]

  generate_secret              = true
  supported_identity_providers = ["COGNITO", "Google"]

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  access_token_validity  = 60
  refresh_token_validity = 30

  depends_on = [aws_cognito_identity_provider.google]
}

resource "aws_cognito_identity_provider" "google" {
  user_pool_id  = aws_cognito_user_pool.pool.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    client_id        = var.google_client_id
    client_secret    = var.google_client_secret
    authorize_scopes = "email openid profile"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}

resource "aws_cognito_user_pool_domain" "main" {
  domain       = "yasuaki640-domain"
  user_pool_id = aws_cognito_user_pool.pool.id
}