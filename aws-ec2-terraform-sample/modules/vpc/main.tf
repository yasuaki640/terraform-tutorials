terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"

      version = "5.99.1"
    }
  }
}
resource "aws_vpc" "main" {
  cidr_block = var.vpc_cidr

  tags = {
    Name = "${var.name_prefix}-vpc"
  }
}

resource "aws_subnet" "public" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.public_subnet_cidr
  availability_zone = var.availability_zone

  # NOTE: 学習目的のため、EC2起動時に自動でパブリックIPを割り当てる設定
  # 本番環境では以下の構成を推奨:
  #   - map_public_ip_on_launch = false
  #   - Elastic IP を EC2 に割り当て
  #   - Route 53 で取得したドメインを Elastic IP に紐付け
  #   - ELB (Application Load Balancer) 経由でアクセス
  # 上記構成により、固定ドメインでアクセス可能なサーバーを構築できる
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.name_prefix}-public-subnet"
  }
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidr
  availability_zone = var.availability_zone

  tags = {
    Name = "${var.name_prefix}-private-subnet"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.name_prefix}-igw"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.name_prefix}-public-rtb"
  }
}

resource "aws_route" "public_internet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}