# =============================================================================
# Cognito User Pool
# =============================================================================
#
# なぜ Cognito User Pool が必要か:
# OIDC/OAuth 2.0 の認証フローには「認可サーバー」が必要。
# Cognito User Pool はマネージドな認可サーバーとして機能し、
# ユーザー管理・認証・トークン発行を一元的に担当する。
#
# 実務では:
# - マイクロサービスアーキテクチャで共通の認証基盤として利用
# - Hosted UI を使えば、ログイン画面の実装が不要になる
# - 外部 IdP（Google, Facebook 等）との連携も容易

resource "aws_cognito_user_pool" "main" {
  name = var.project_name

  # メールアドレスをユーザー名として使用
  # 実務では: メールアドレスをユーザー名にするのが最も一般的
  username_attributes = ["email"]

  # メール検証を有効化
  # なぜ: ユーザーが本当にそのメールアドレスの所有者かを確認するため
  auto_verified_attributes = ["email"]

  # パスワードポリシー
  # 実務では: NIST SP 800-63B に準拠し、長さ重視のポリシーが推奨される
  password_policy {
    minimum_length                   = 8
    require_lowercase                = true
    require_uppercase                = true
    require_numbers                  = true
    require_symbols                  = true
    temporary_password_validity_days = 7
  }

  # セルフサインアップを許可
  # 注意: 本番環境では管理者のみがユーザーを作成する設定も検討すること
  admin_create_user_config {
    allow_admin_create_user_only = false
  }

  # メール設定（Cognito デフォルトのメール送信を使用）
  # 実務では: SES を使って独自ドメインからメールを送信する
  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  # スキーマ属性
  schema {
    name                     = "email"
    attribute_data_type      = "String"
    required                 = true
    mutable                  = true
    developer_only_attribute = false

    string_attribute_constraints {
      min_length = 1
      max_length = 256
    }
  }

  tags = {
    Project     = var.project_name
    Environment = "learning"
    ManagedBy   = "terraform"
  }
}

# =============================================================================
# Cognito User Pool Client（App Client）
# =============================================================================
#
# なぜ App Client が必要か:
# OAuth 2.0 では、アプリケーション（クライアント）を認可サーバーに登録する必要がある。
# App Client = OAuth 2.0 の「クライアント登録」に相当。
#
# PKCE を使う理由:
# SPA やモバイルアプリなどの「公開クライアント」では、
# client_secret を安全に保管できない。
# PKCE（Proof Key for Code Exchange）は client_secret の代わりに
# 動的に生成した code_verifier / code_challenge を使って
# 認可コード横取り攻撃を防ぐ（RFC 7636）。

resource "aws_cognito_user_pool_client" "main" {
  name         = "${var.project_name}-client"
  user_pool_id = aws_cognito_user_pool.main.id

  # クライアントシークレットなし（公開クライアント向け）
  # なぜ: SPA/モバイルアプリでは client_secret を安全に保管できないため
  # 代わりに PKCE で認可コード横取り攻撃を防ぐ
  generate_secret = false

  # 許可する OAuth フロー: Authorization Code Grant のみ
  # なぜ code だけか:
  # - implicit: トークンが URL フラグメントに露出するため非推奨（OAuth 2.1 で廃止予定）
  # - code: 認可コード→トークン交換の2段階により、トークンの露出リスクを最小化
  allowed_oauth_flows = ["code"]

  # 許可するスコープ
  # - openid: OIDC 必須（ID Token を取得するために必要）
  # - email: メールアドレス関連の claims を取得
  # - profile: プロフィール情報の claims を取得
  allowed_oauth_scopes = ["openid", "email", "profile"]

  # OAuth フロー設定を有効化（これがないと allowed_oauth_flows が無視される）
  allowed_oauth_flows_user_pool_client = true

  # コールバック URL（認可コードが送られる先）
  callback_urls = [var.callback_url]

  # ログアウト URL
  logout_urls = [var.logout_url]

  # サポートする ID プロバイダー
  # COGNITO = Cognito User Pool 自体をプロバイダーとして使用
  # 実務では: Google, Facebook 等の外部 IdP を追加することも多い
  supported_identity_providers = ["COGNITO"]

  # トークンの有効期限設定
  # 学習ポイント: 各トークンの役割と有効期限の考え方
  # - Access Token: APIアクセスの「認可」に使う。短めに設定（漏洩時のリスク軽減）
  # - ID Token: ユーザーの「認証」情報を含む。Access Token と同じ期間が一般的
  # - Refresh Token: 新しい Access/ID Token を取得するために使う。長めに設定
  access_token_validity  = 1  # 1時間
  id_token_validity      = 1  # 1時間
  refresh_token_validity = 30 # 30日

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  # トークン取り消し（Revocation）を有効化
  enable_token_revocation = true

  # CSRF 対策: state パラメーターの検証を防ぐ
  # （Cognito 側では state は透過的に扱われる）
  prevent_user_existence_errors = "ENABLED"
}

# =============================================================================
# Cognito User Pool Domain（Hosted UI）
# =============================================================================
#
# なぜ Hosted UI ドメインが必要か:
# Cognito Hosted UI は、ログイン/サインアップ/パスワードリセット画面を
# マネージドに提供する。独自のログイン画面を実装する手間を省ける。
#
# 実務では:
# - 本番環境ではカスタムドメイン（例: auth.example.com）を使用する
# - Hosted UI のデザインはカスタマイズ可能だが、制約も多い
# - 完全にカスタマイズしたい場合は、Hosted UI を使わず自前で実装する

resource "aws_cognito_user_pool_domain" "main" {
  domain       = var.project_name
  user_pool_id = aws_cognito_user_pool.main.id
}
