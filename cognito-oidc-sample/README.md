# Cognito OIDC Sample - 手動 PKCE + JWT 署名検証

## このプロジェクトについて

OIDC（OpenID Connect）の Authorization Code + PKCE フローを**外部ライブラリなし**で実装した学習用サンプルです。

Go の標準ライブラリとブラウザの Web Crypto API のみを使用し、以下を手動で実装しています:

- **PKCE（Proof Key for Code Exchange）**: `code_verifier`/`code_challenge` の生成と検証
- **JWT 署名検証**: JWKS からの公開鍵取得、RSA-SHA256 署名の検証
- **トークン交換**: Cognito の `/oauth2/token` エンドポイントとの通信

## `aws-cognito-sample` との違い

| 項目 | aws-cognito-sample | cognito-oidc-sample（本プロジェクト） |
|------|-------------------|--------------------------------------|
| Go 外部依存 | go-oidc, oauth2, gorilla, godotenv, golang-jwt | **なし（標準ライブラリのみ）** |
| JWT 検証 | `ParseUnverified`（署名検証なし） | **RSA-SHA256 手動署名検証** |
| PKCE | なし | **フロントエンドで手動実装** |
| クライアント種別 | Confidential（シークレットあり） | **Public（シークレットなし）** |
| フロントエンド | Go テンプレート（SSR） | **Vanilla JS SPA** |
| セッション管理 | gorilla/sessions（Cookie） | **トークンベース（サーバーサイドセッションなし）** |
| Docker | なし | **docker-compose + マルチステージビルド** |
| AWS Provider | ~> 4.16 | **~> 5.0** |

## 前提条件

- [AWS CLI](https://aws.amazon.com/cli/) 設定済み
- [Terraform](https://www.terraform.io/) >= 1.0
- [Docker](https://www.docker.com/) & Docker Compose
- [Go](https://golang.org/) >= 1.22（ローカル実行時のみ）

## アーキテクチャ

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   ブラウザ       │     │   Go バックエンド  │     │   Cognito       │
│   (Vanilla JS)  │     │   (stdlib only)   │     │                 │
│                 │     │                   │     │                 │
│  PKCE 生成      │────>│  POST /callback   │────>│  /oauth2/token  │
│  認可リクエスト   │     │  (トークン交換)    │<────│  (トークン発行)  │
│                 │<────│                   │     │                 │
│  GET /me        │────>│  JWT 署名検証      │────>│  JWKS 公開鍵    │
│  (Bearer Token) │<────│  (RSA-SHA256)     │     │                 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
     同一オリジン (localhost:8080) → CORS 不要
```

## セットアップ

### 1. Terraform でインフラ構築

```bash
# Terraform 変数の設定
cd infra
cp terraform.tfvars.example terraform.tfvars
# 必要に応じて project_name を変更（AWS全体で一意である必要あり）

# 初期化 & リソース作成
make infra-init
make infra-apply
```

### 2. 環境変数の生成

```bash
# Terraform output から .env ファイルを自動生成
make env
```

### 3. アプリケーション起動

```bash
# Docker Compose で起動
make docker-up

# ログの確認
make docker-logs
```

### 4. ブラウザでアクセス

http://localhost:8080/ を開き、「Cognito でログイン」ボタンをクリック。
初回はユーザー登録（Sign Up）が必要です。

## Makefile コマンド一覧

| コマンド | 説明 |
|---------|------|
| `make infra-init` | Terraform 初期化 |
| `make infra-plan` | 変更プレビュー |
| `make infra-apply` | リソース作成 |
| `make infra-destroy` | リソース削除 |
| `make env` | .env ファイル自動生成 |
| `make run` | ローカル実行（Docker不要） |
| `make docker-up` | Docker Compose 起動 |
| `make docker-down` | Docker Compose 停止 |
| `make docker-logs` | ログ確認 |
| `make clean` | Docker + .env クリーンアップ |

## 学習ポイント

### 1. PKCE（Proof Key for Code Exchange）
`frontend/index.html` の `generateCodeVerifier()` と `generateCodeChallenge()` を参照。
Web Crypto API で SHA-256 ハッシュを計算し、Base64URL エンコードする。

### 2. JWT 署名検証
`backend/pkg/token/verify.go` を参照。
JWT の3パート分割、ヘッダー解析、RSA 署名検証、クレーム検証の全手順を手動実装。

### 3. JWKS 公開鍵の取得とキャッシュ
`backend/pkg/token/jwks.go` を参照。
JWK の N（modulus）と E（exponent）から RSA 公開鍵を構築する過程を学べる。

### 4. Authorization Code フロー全体
`docs/flow.md` にシーケンス図と各ステップの詳細解説あり。

## セキュリティに関する注意

- **localhost 使用**: 学習用のため HTTP を使用。本番では必ず HTTPS を使用すること。
- **トークン保存**: メモリ変数に保存（XSS 対策）。本番では `httpOnly` Cookie を推奨。
- **`ParseUnverified` の危険性**: 既存サンプルでは署名検証をスキップしている。本プロジェクトでは正しく RSA 署名検証を実装している。
- **CORS**: 同一オリジン配信のため不要。フロントエンドとバックエンドを分離する場合は適切な CORS 設定が必要。
- **Cognito ドメイン**: 本番ではカスタムドメインを設定し、フィッシング対策を行うこと。

## クリーンアップ

```bash
# Docker リソースの削除
make clean

# AWS リソースの削除（課金を止める）
make infra-destroy
```

Cognito User Pool 自体に時間課金はありませんが、不要なリソースは削除する習慣をつけましょう。
