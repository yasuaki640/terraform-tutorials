# cognito-oidc-sample

OIDC（OpenID Connect）/ OAuth 2.0 の学習用サンプルプロジェクト。
AWS Cognito を Identity Provider として、**Authorization Code Flow + PKCE** を Go バックエンド + 素の HTML/JS フロントエンドで実装する。

## 学習目標

- Authorization Code Flow + PKCE の各ステップを手で実装し、仕様を体感する
- ID Token / Access Token / Refresh Token の役割の違いを理解する
- JWKS による JWT 署名検証を Go で実装する

## 構成

```
cognito-oidc-sample/
├── infra/          # Terraform（Cognito User Pool + App Client）
├── frontend/       # 素の HTML/JS（PKCE 生成 + コールバック処理）
├── backend/        # Go（トークン交換 + JWT 検証）
└── docs/           # OIDC フロー図解メモ
```

## 前提条件

- AWS アカウント + 適切な IAM 権限
- Terraform >= 1.0
- Go >= 1.22
- Docker & Docker Compose（任意）

## セットアップ

### 1. Cognito 構築（Terraform）

```bash
make infra-init   # terraform init
make infra-plan   # terraform plan（確認）
make infra-apply  # terraform apply（リソース作成）
```

### 2. 環境変数の設定

```bash
# terraform output から .env を自動生成
make env

# または手動で .env.example をコピーして編集
cp .env.example .env
```

### 3. アプリケーション起動

```bash
# Go 直接実行
make run

# または Docker
make docker-up
```

### 4. ブラウザでアクセス

http://localhost:8080 にアクセスし、「Login with Cognito」ボタンをクリック。

## 技術スタック

| 技術 | 用途 |
|------|------|
| Terraform | Cognito インフラ構築 |
| Go (net/http) | バックエンド API（フレームワーク不使用） |
| HTML/JS | フロントエンド（ライブラリ不使用） |
| AWS Cognito | OIDC Identity Provider |

## 注意事項

- このサンプルは**学習用**であり、本番利用を想定していない
- トークンの保存は `sessionStorage` を使用（本番では HttpOnly Cookie 推奨）
- CORS 設定はローカル開発用に緩く設定している
- Terraform の state はローカル管理（S3 backend は不要）
