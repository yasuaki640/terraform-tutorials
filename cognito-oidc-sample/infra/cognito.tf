# =============================================================================
# Amazon Cognito User Pool
# =============================================================================
# なぜ Cognito を使うか:
#   ユーザー認証基盤をゼロから構築するのは非常に困難（パスワードハッシュ、MFA、
#   メール検証、ブルートフォース対策など）。Cognito はこれらをマネージドサービス
#   として提供し、OIDC/OAuth 2.0 標準に準拠したエンドポイントを自動生成する。
#
# aws-cognito-sample との違い:
#   既存サンプルでは go-oidc ライブラリがトークン検証を行うため、内部の仕組みが
#   隠蔽されていた。このサンプルではライブラリを使わず、OIDC の各ステップを
#   手動で実装することで、プロトコルの理解を深める。
# =============================================================================

resource "aws_cognito_user_pool" "main" {
  name = var.project_name

  # メールアドレスをユーザー名として使用
  # なぜ: ユーザーにとって覚えやすく、一意性が保証される
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  # パスワードポリシー
  # なぜ厳格にするか: セキュリティの基本原則。弱いパスワードは攻撃者にとって
  # 最も簡単な侵入経路。実務では組織のセキュリティポリシーに従う。
  # 学習用でも本番に近い設定を体験しておくことが重要。
  password_policy {
    minimum_length                   = 8
    require_lowercase                = true
    require_uppercase                = true
    require_numbers                  = true
    require_symbols                  = true
    temporary_password_validity_days = 7
  }

  # メール検証設定
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  # アカウントリカバリー設定
  # なぜ email を最優先にするか: SMS は SIM スワップ攻撃のリスクがある。
  # メールベースのリカバリーがより安全。
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }
}

# =============================================================================
# App Client（パブリッククライアント）
# =============================================================================
# ★ このサンプルの最重要ポイント ★
#
# 【aws-cognito-sample との決定的な違い】
#   既存サンプル: generate_secret = true （Confidential Client）
#     → サーバーサイドにクライアントシークレットを保持
#     → シークレット + 認可コードでトークンを取得
#
#   このサンプル: generate_secret = false （Public Client）
#     → クライアントシークレットなし
#     → 代わりに PKCE（Proof Key for Code Exchange）でコードを保護
#
# 【なぜ Public Client + PKCE なのか】
#   SPA（Single Page Application）やモバイルアプリは、クライアントシークレットを
#   安全に保持できない（ブラウザの DevTools やアプリのリバースエンジニアリングで
#   容易に漏洩する）。PKCE は認可リクエスト時に一時的な検証コード（code_verifier）
#   を生成し、そのハッシュ（code_challenge）を送信する。トークン交換時に元の
#   code_verifier を提示することで、認可コードの横取り攻撃を防ぐ。
#
#   RFC 7636: https://tools.ietf.org/html/rfc7636
# =============================================================================

resource "aws_cognito_user_pool_client" "main" {
  name         = "${var.project_name}-client"
  user_pool_id = aws_cognito_user_pool.main.id

  # パブリッククライアント: クライアントシークレットを生成しない
  generate_secret = false

  # 認可コードフローのみ許可
  # なぜ Implicit Flow を使わないか:
  #   Implicit Flow（response_type=token）はアクセストークンがURLフラグメントに
  #   含まれるため、ブラウザ履歴やリファラーヘッダーで漏洩するリスクがある。
  #   OAuth 2.0 Security Best Current Practice (RFC 9700) で非推奨とされている。
  #   Authorization Code + PKCE が現在のベストプラクティス。
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["openid", "email", "profile"]

  # コールバックURL / ログアウトURL
  # 学習用: localhost を使用。本番では HTTPS の実ドメインを指定する。
  callback_urls = [var.callback_url]
  logout_urls   = [var.logout_url]

  # トークン有効期限
  # なぜ短くするか: アクセストークンが漏洩した場合の影響範囲を限定する。
  # 実務では要件に応じて調整するが、短いほどセキュア。
  access_token_validity  = 1  # 1時間
  id_token_validity      = 1  # 1時間
  refresh_token_validity = 30 # 30日

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  # ID プロバイダー
  # このサンプルでは Cognito 自身のみ。Google 連携は aws-cognito-sample を参照。
  supported_identity_providers = ["COGNITO"]

  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH",
  ]
}

# =============================================================================
# Cognito Hosted UI ドメイン
# =============================================================================
# なぜ Hosted UI を使うか:
#   ログインフォームを自前で作ると、パスワードの安全な送信、CSRF対策、
#   MFA対応など多くのセキュリティ要件を満たす必要がある。
#   Hosted UI はこれらを自動で処理し、OIDCの認可エンドポイント（/authorize）
#   とトークンエンドポイント（/oauth2/token）を提供する。
#
# 注意: ドメインプレフィックスは AWS 全体でグローバルに一意である必要がある。
#   "cognito-oidc-sample" が他のアカウントで使用済みの場合は、
#   terraform.tfvars で project_name を変更すること。
# =============================================================================

resource "aws_cognito_user_pool_domain" "main" {
  domain       = var.project_name
  user_pool_id = aws_cognito_user_pool.main.id
}
