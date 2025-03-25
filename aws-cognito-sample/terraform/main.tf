terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
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
}

resource "aws_cognito_user_pool_client" "client" {
  name         = "aws-cognito-sample-application-client"
  user_pool_id = aws_cognito_user_pool.pool.id

  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["email", "openid", "profile"]

  callback_urls        = ["http://localhost/callback"]
  logout_urls          = ["http://localhost/logout"]
  default_redirect_uri = "http://localhost/callback"

  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_PASSWORD_AUTH"
  ]

  generate_secret              = true
  supported_identity_providers = ["COGNITO"]

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  access_token_validity  = 60
  id_token_validity      = 60
  refresh_token_validity = 30
}

resource "aws_cognito_user_pool_client" "m2m_client" {
  name         = "aws-cognito-sample-m2m-client"
  user_pool_id = aws_cognito_user_pool.pool.id

  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes = [
    "email",
    "openid",
    "profile",
    "default-m2m-resource-server-d7rdxs/read"
  ]

  generate_secret              = true
  supported_identity_providers = ["COGNITO"]

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  access_token_validity  = 60
  id_token_validity      = 60
  refresh_token_validity = 30
}


resource "aws_cognito_user_pool_domain" "main" {
  domain       = "yasuaki640-domain"
  user_pool_id = aws_cognito_user_pool.pool.id
}