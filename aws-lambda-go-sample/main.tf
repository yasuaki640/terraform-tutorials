terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"
}

# Lambda関数の実行ロール
resource "aws_iam_role" "lambda_role" {
  name = "lambda_execution_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# CloudWatch Logsのポリシーをアタッチ
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda関数
resource "aws_lambda_function" "go_lambda" {
  filename         = "function.zip"
  function_name    = "go-lambda-function"
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap" # mainからbootstrapに変更
  runtime         = "provided.al2" # go1.xからprovided.al2に変更
  source_code_hash = filebase64sha256("function.zip")
  publish          = true # 新しいバージョンを発行
}

resource "aws_lambda_alias" "live_alias" {
  name             = "live"
  description      = "Points to the latest published version"
  function_name    = aws_lambda_function.go_lambda.function_name
  function_version = aws_lambda_function.go_lambda.version
}
