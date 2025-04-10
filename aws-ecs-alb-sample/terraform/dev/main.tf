provider "aws" {
  region = "ap-northeast-1"
  allowed_account_ids = var.allowed_account_ids
}

terraform {
  backend "s3" {
    bucket         = "terraform-state-yasuaki640-ecs-alb"
    key            = "aws-ecs-alb-sample/dev/terraform.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "terraform-state-locks"
    encrypt        = true
  }
}
