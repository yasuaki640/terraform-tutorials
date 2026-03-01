package handler

import (
	"net/http"
)

// =============================================================================
// 設定ハンドラー
// =============================================================================
// フロントエンドが OIDC 認可リクエストを構築するために必要な設定値を返す。
//
// なぜこのエンドポイントが必要か:
//   フロントエンドは静的な HTML/JS ファイルとして配信されるため、
//   環境変数やサーバー設定に直接アクセスできない。
//   このエンドポイントを通じて、Cognito ドメインやクライアント ID を
//   動的に取得する。
//
//   代替手段:
//   - Go テンプレートで HTML に埋め込む → FileServer が使えなくなる
//   - JS に直接ハードコード → 環境変更時に HTML を編集する必要がある
//   - GET /config は最もクリーンで、SPAでは一般的なパターン
// =============================================================================

// Config はフロントエンド向けの設定を返すハンドラー。
type Config struct {
	cognitoDomain string
	clientID      string
	redirectURI   string
}

// NewConfig は新しい Config ハンドラーを作成する。
func NewConfig(cognitoDomain, clientID, redirectURI string) *Config {
	return &Config{
		cognitoDomain: cognitoDomain,
		clientID:      clientID,
		redirectURI:   redirectURI,
	}
}

// configResponse はフロントエンドに返す設定値。
type configResponse struct {
	CognitoDomain string `json:"cognito_domain"`
	ClientID      string `json:"client_id"`
	RedirectURI   string `json:"redirect_uri"`
}

// Handle は GET /config のリクエストを処理する。
func (c *Config) Handle(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, configResponse{
		CognitoDomain: c.cognitoDomain,
		ClientID:      c.clientID,
		RedirectURI:   c.redirectURI,
	})
}
