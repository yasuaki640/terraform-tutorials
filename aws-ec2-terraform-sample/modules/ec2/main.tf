terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.99.1"
    }
  }
}

# TODO(human): Security Group for EC2
resource "aws_security_group" "web" {
  name        = "${var.name_prefix}-web-sg"
  description = "Security group for web server"
  # TODO(human): vpc_id を追加してください

  tags = {
    Name = "${var.name_prefix}-web-sg"
  }
}

# TODO(human): Security Group Rules
# ここにセキュリティグループルールを定義してください
# - HTTPアクセス用(port 80)のインバウンドルール
# - SSH接続用(port 22)のインバウンドルール
# - すべてのアウトバウンドトラフィックを許可するルール

# TODO(human): EC2 Instance
# ここにEC2インスタンスリソースを定義してください