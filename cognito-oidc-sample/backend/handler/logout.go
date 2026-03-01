package handler

import (
	"net/http"
	"net/url"
)

// =============================================================================
// ログアウトハンドラー
// =============================================================================
// Cognito のログアウトフロー:
//   1. フロントエンドが POST /logout を呼び出す
//   2. バックエンドが Cognito のログアウト URL を構築して返す
//   3. フロントエンドがその URL にリダイレクト
//   4. Cognito がセッションを破棄し、logout_uri にリダイレクト
//
// なぜフロントエンドから直接リダイレクトしないか:
//   技術的には可能だが、ログアウト URL の構築に必要な設定値
//   （cognito_domain, client_id, logout_uri）をフロントエンドに
//   持たせない方がクリーン。GET /config で取得した値を使ってフロントエンドで
//   構築することもできるが、バックエンドに集約する方が管理しやすい。
// =============================================================================

// Logout はログアウト URL を返すハンドラー。
type Logout struct {
	cognitoDomain    string
	clientID         string
	logoutRedirectURI string
}

// NewLogout は新しい Logout ハンドラーを作成する。
func NewLogout(cognitoDomain, clientID, logoutRedirectURI string) *Logout {
	return &Logout{
		cognitoDomain:    cognitoDomain,
		clientID:         clientID,
		logoutRedirectURI: logoutRedirectURI,
	}
}

// logoutResponse はフロントエンドに返すレスポンス。
type logoutResponse struct {
	LogoutURL string `json:"logout_url"`
}

// Handle は POST /logout のリクエストを処理する。
func (l *Logout) Handle(w http.ResponseWriter, r *http.Request) {
	// Cognito ログアウト URL の構築
	// https://docs.aws.amazon.com/cognito/latest/developerguide/logout-endpoint.html
	logoutURL := l.cognitoDomain + "/logout?" + url.Values{
		"client_id":  {l.clientID},
		"logout_uri": {l.logoutRedirectURI},
	}.Encode()

	writeJSON(w, http.StatusOK, logoutResponse{
		LogoutURL: logoutURL,
	})
}
