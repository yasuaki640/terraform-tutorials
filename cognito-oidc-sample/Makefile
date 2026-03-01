.PHONY: infra-init infra-plan infra-apply env run docker-up clean

# Terraform コマンド
infra-init:
	cd infra && terraform init

infra-plan:
	cd infra && terraform plan

infra-apply:
	cd infra && terraform apply

# terraform output から .env ファイルを生成
env:
	@echo "# Cognito設定（terraform output から自動生成）" > .env
	@echo "COGNITO_REGION=$$(cd infra && terraform output -raw aws_region 2>/dev/null || echo 'ap-northeast-1')" >> .env
	@echo "COGNITO_USER_POOL_ID=$$(cd infra && terraform output -raw user_pool_id)" >> .env
	@echo "COGNITO_CLIENT_ID=$$(cd infra && terraform output -raw client_id)" >> .env
	@echo "COGNITO_DOMAIN=$$(cd infra && terraform output -raw cognito_domain)" >> .env
	@echo "COGNITO_JWKS_URI=$$(cd infra && terraform output -raw jwks_uri)" >> .env
	@echo "COGNITO_ISSUER=$$(cd infra && terraform output -raw issuer)" >> .env
	@echo "CALLBACK_URL=http://localhost:8080/callback" >> .env
	@echo "FRONTEND_URL=http://localhost:3000" >> .env
	@echo "PORT=8080" >> .env
	@echo ".env ファイルを生成しました"

# ローカル実行（Go直接）
run:
	cd backend && go run main.go

# Docker 実行
docker-up:
	docker compose up --build

clean:
	docker compose down
