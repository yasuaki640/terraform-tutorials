# =============================================================================
# Makefile - cognito-oidc-sample
# =============================================================================
# なぜ Makefile を使うか:
#   複数のコマンドを覚える必要がなく、`make xxx` で操作できる。
#   チーム開発では「どのコマンドを実行すればよいか」の共通言語になる。
# =============================================================================

.PHONY: infra-init infra-plan infra-apply infra-destroy env run docker-up docker-down docker-logs clean

# --- Terraform 操作 ---

## Terraform の初期化（プロバイダーのダウンロード等）
infra-init:
	cd infra && terraform init

## 変更内容のプレビュー（実際のリソース変更なし）
infra-plan:
	cd infra && terraform plan

## リソースの作成・更新
infra-apply:
	cd infra && terraform apply

## リソースの削除（コスト削減のため、使い終わったら必ず実行）
infra-destroy:
	cd infra && terraform destroy

# --- 環境変数の生成 ---

## Terraform output から .env ファイルを自動生成
env:
	@echo "COGNITO_DOMAIN=$$(cd infra && terraform output -raw cognito_domain)" > .env
	@echo "CLIENT_ID=$$(cd infra && terraform output -raw client_id)" >> .env
	@echo "REDIRECT_URI=http://localhost:8080/callback" >> .env
	@echo "LOGOUT_REDIRECT_URI=http://localhost:8080" >> .env
	@echo "JWKS_URI=$$(cd infra && terraform output -raw jwks_uri)" >> .env
	@echo "ISSUER=$$(cd infra && terraform output -raw issuer)" >> .env
	@echo ".env ファイルを生成しました"
	@cat .env

# --- ローカル実行（Docker 不要） ---

## Go サーバーを直接実行（.env の手動読み込みが必要）
run:
	cd backend && go run .

# --- Docker 操作 ---

## Docker Compose でビルド＆起動
docker-up:
	docker compose up --build -d

## Docker Compose で停止
docker-down:
	docker compose down

## ログの確認
docker-logs:
	docker compose logs -f

# --- クリーンアップ ---

## Docker リソースと .env を削除
clean:
	docker compose down --rmi local --volumes 2>/dev/null || true
	rm -f .env
	rm -f backend/backend
