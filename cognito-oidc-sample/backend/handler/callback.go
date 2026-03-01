// Package handler は HTTP リクエストハンドラーを提供する。
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// =============================================================================
// コールバックハンドラー（トークン交換）
// =============================================================================
// このハンドラーは OIDC Authorization Code Flow の最終ステップを担当する。
//
// フロントエンドから受け取るもの:
//   - code: Cognito が発行した認可コード（短命・一度きり使用可能）
//   - code_verifier: PKCE のコード検証子（フロントエンドで生成）
//
// このハンドラーが行うこと:
//   Cognito の /oauth2/token エンドポイントに POST リクエストを送信し、
//   認可コードをアクセストークン・IDトークン・リフレッシュトークンに交換する。
//
// なぜバックエンドでトークン交換するか:
//   フロントエンド（ブラウザ）から直接 /oauth2/token を呼ぶことも技術的には
//   可能だが、バックエンドを経由することで:
//   - トークンエンドポイントの URL やパラメーターをクライアントから隠蔽
//   - 将来的にトークンのフィルタリングや追加検証を挟める
//   - Confidential Client への移行が容易
//
// PKCE（Proof Key for Code Exchange）の仕組み:
//   1. フロントエンドが code_verifier（ランダム文字列）を生成
//   2. code_challenge = SHA256(code_verifier) を計算（Base64URL）
//   3. /authorize に code_challenge を送信
//   4. Cognito がコールバックで返す code は code_challenge と紐づく
//   5. /oauth2/token に code + code_verifier を送信
//   6. Cognito が SHA256(code_verifier) == code_challenge を検証
//   → 認可コードを横取りした攻撃者は code_verifier を持たないため、
//     トークン交換できない
// =============================================================================

// Callback はトークン交換を行うハンドラー。
type Callback struct {
	cognitoDomain string
	clientID      string
	redirectURI   string
}

// NewCallback は新しい Callback ハンドラーを作成する。
func NewCallback(cognitoDomain, clientID, redirectURI string) *Callback {
	return &Callback{
		cognitoDomain: cognitoDomain,
		clientID:      clientID,
		redirectURI:   redirectURI,
	}
}

// callbackRequest はフロントエンドからのリクエストボディ。
type callbackRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

// tokenResponse は Cognito のトークンエンドポイントからのレスポンス。
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Handle は POST /callback のリクエストを処理する。
func (c *Callback) Handle(w http.ResponseWriter, r *http.Request) {
	// リクエストボディのパース
	var req callbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "リクエストボディのパースに失敗しました",
		})
		return
	}

	if req.Code == "" || req.CodeVerifier == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "code と code_verifier は必須です",
		})
		return
	}

	// Cognito トークンエンドポイントへの POST リクエストを構築
	// Content-Type: application/x-www-form-urlencoded（OAuth 2.0 仕様）
	tokenURL := c.cognitoDomain + "/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("code", req.Code)
	data.Set("redirect_uri", c.redirectURI)
	data.Set("code_verifier", req.CodeVerifier) // PKCE: code_verifier を送信

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("トークンエンドポイントへのリクエスト失敗: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "トークンの取得に失敗しました",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("レスポンスボディの読み取り失敗: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "トークンレスポンスの読み取りに失敗しました",
		})
		return
	}

	// Cognito がエラーを返した場合
	if resp.StatusCode != http.StatusOK {
		log.Printf("トークンエンドポイントがエラーを返しました: status=%d, body=%s", resp.StatusCode, string(body))
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("トークンの取得に失敗しました（Cognito: %d）", resp.StatusCode),
		})
		return
	}

	// Cognito のレスポンスをそのままフロントエンドに返す
	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("トークンレスポンスのパース失敗: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "トークンレスポンスのパースに失敗しました",
		})
		return
	}

	writeJSON(w, http.StatusOK, tokenResp)
}

// writeJSON は JSON レスポンスを書き込むヘルパー関数。
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
