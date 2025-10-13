terraform {
  backend "s3" {
    bucket         = "terraform-state-839063654285-ap-northeast-1"
    key            = "dev/terraform.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}