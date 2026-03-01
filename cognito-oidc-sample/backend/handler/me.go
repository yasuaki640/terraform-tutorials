package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yasuaki640/cognito-oidc-sample/backend/middleware"
)

// MeHandler は認証済みユーザーの情報を返すハンドラー。
//
// 学習ポイント: 実運用では /me の情報は ID Token から取得するのが一般的。
// Access Token の claims にはユーザー属性が含まれない場合がある。
// ここでは Access Token で「認可」し、必要なら Cognito の
// /oauth2/userInfo エンドポイントからユーザー属性を取得するパターンも考えられる。
type MeHandler struct{}

type meResponse struct {
	Sub      string `json:"sub"`
	Email    string `json:"email,omitempty"`
	TokenUse string `json:"token_use"`
}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// context から claims を取得（auth ミドルウェアがセット済み）
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "認証情報が見つかりません", http.StatusUnauthorized)
		return
	}

	resp := meResponse{
		Sub:      claims.Sub,
		Email:    claims.Email,
		TokenUse: claims.TokenUse,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
