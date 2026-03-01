package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// LogoutHandler は Cognito のログアウト URL を生成して返すハンドラー。
//
// 学習ポイント: OAuth 2.0 / OIDC にはログアウトの標準仕様がいくつかある:
// - RP-Initiated Logout（OpenID Connect RP-Initiated Logout 1.0）
// - Cognito 独自の /logout エンドポイント
// ここでは Cognito 独自のエンドポイントを使用する。
// クライアント側で Cognito の /logout にリダイレクトすることで、
// Cognito 側のセッション（Cookie）も破棄される。
type LogoutHandler struct {
	cognitoDomain string
	clientID      string
	logoutURL     string
}

type logoutResponse struct {
	LogoutURL string `json:"logout_url"`
}

func NewLogoutHandler(cognitoDomain, clientID, logoutURL string) *LogoutHandler {
	return &LogoutHandler{
		cognitoDomain: cognitoDomain,
		clientID:      clientID,
		logoutURL:     logoutURL,
	}
}

func (h *LogoutHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Cognito の /logout エンドポイント URL を生成
	// パラメーター:
	// - client_id: どのアプリクライアントからのログアウトか
	// - logout_uri: ログアウト後のリダイレクト先（Cognito に登録済みの URL）
	params := url.Values{
		"client_id":  {h.clientID},
		"logout_uri": {h.logoutURL},
	}

	logoutEndpoint := h.cognitoDomain + "/logout?" + params.Encode()

	resp := logoutResponse{
		LogoutURL: logoutEndpoint,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
