package handler

import (
	"net/http"

	"github.com/yasuaki640/terraform-tutorials/cognito-oidc-sample/backend/middleware"
)

// =============================================================================
// ユーザー情報ハンドラー
// =============================================================================
// このハンドラーは認証ミドルウェアで検証済みの Access Token から
// ユーザー情報を取得して返す。
//
// なぜ Cognito の /userInfo エンドポイントを使わないか:
//   Access Token のクレームに必要な情報（sub, email 等）が含まれているため、
//   追加の HTTP リクエストなしでユーザー情報を返せる。
//   実務で追加の属性が必要な場合は /userInfo エンドポイントを使う。
// =============================================================================

// Me はユーザー情報を返すハンドラー。
type Me struct{}

// NewMe は新しい Me ハンドラーを作成する。
func NewMe() *Me {
	return &Me{}
}

// meResponse はクライアントに返すユーザー情報。
type meResponse struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Handle は GET /me のリクエストを処理する。
// 認証ミドルウェアを通過済みのため、context に Claims が格納されている。
func (m *Me) Handle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		// ミドルウェアを通過していれば到達しないが、安全のため
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "認証情報がありません",
		})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		Sub:      claims.Sub,
		Email:    claims.Email,
		Username: claims.Username,
	})
}
