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

// CallbackHandler は認可コードをトークンに交換するハンドラー。
//
// 学習ポイント: Authorization Code Flow の2番目のステップ。
// フロントエンドから受け取った認可コードと code_verifier を使って、
// Cognito のトークンエンドポイントにリクエストし、
// ID Token / Access Token / Refresh Token を取得する。
type CallbackHandler struct {
	cognitoDomain string
	clientID      string
	callbackURL   string
	httpClient    *http.Client
}

// callbackRequest はフロントエンドからのリクエストボディ
type callbackRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

// tokenResponse は Cognito トークンエンドポイントからのレスポンス
type tokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewCallbackHandler(cognitoDomain, clientID, callbackURL string) *CallbackHandler {
	return &CallbackHandler{
		cognitoDomain: cognitoDomain,
		clientID:      clientID,
		callbackURL:   callbackURL,
		httpClient:    &http.Client{},
	}
}

func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. リクエストボディをパース
	var req callbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストボディのパースに失敗", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.CodeVerifier == "" {
		http.Error(w, "code と code_verifier は必須です", http.StatusBadRequest)
		return
	}

	log.Printf("トークン交換開始: code=%s...", req.Code[:min(len(req.Code), 10)])

	// 2. Cognito のトークンエンドポイントにリクエスト
	//
	// 学習ポイント: トークンエンドポイントへのリクエストは
	// application/x-www-form-urlencoded（JSON ではない）。
	// これは OAuth 2.0 仕様（RFC 6749 Section 4.1.3）に準拠。
	tokenURL := fmt.Sprintf("%s/oauth2/token", h.cognitoDomain)

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {h.clientID},
		"code":          {req.Code},
		"redirect_uri":  {h.callbackURL},
		"code_verifier": {req.CodeVerifier},
	}

	tokenReq, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("トークンリクエスト作成失敗: %v", err)
		http.Error(w, "内部エラー", http.StatusInternalServerError)
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 3. Cognito にリクエスト送信
	resp, err := h.httpClient.Do(tokenReq)
	if err != nil {
		log.Printf("トークンエンドポイントへのリクエスト失敗: %v", err)
		http.Error(w, "トークンエンドポイントへの接続に失敗", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("レスポンス読み取り失敗: %v", err)
		http.Error(w, "内部エラー", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("トークン交換失敗: status=%d, body=%s", resp.StatusCode, string(body))
		http.Error(w, fmt.Sprintf("トークン交換失敗: %s", string(body)), http.StatusBadRequest)
		return
	}

	// 4. レスポンスをパースしてクライアントに返却
	//
	// 学習ポイント: Cognito からのレスポンスには以下の3つのトークンが含まれる:
	// - id_token: ユーザーの「認証」情報（誰であるか）を含む JWT
	// - access_token: API アクセスの「認可」に使う JWT
	// - refresh_token: 新しい access/id token を取得するためのトークン（JWT ではない）
	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("トークンレスポンスのパースに失敗: %v", err)
		http.Error(w, "内部エラー", http.StatusInternalServerError)
		return
	}

	log.Printf("トークン交換成功")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp)
}
