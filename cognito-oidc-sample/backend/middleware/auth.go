// Package middleware は HTTP ミドルウェアを提供する。
//
// ミドルウェアとは:
//   ハンドラーの前後に共通の処理を挟む仕組み。認証・ログ・CORS等を
//   各ハンドラーから分離し、再利用可能にする。
//   Go では http.Handler をラップする関数として実装するのが一般的。
package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/yasuaki640/terraform-tutorials/cognito-oidc-sample/backend/pkg/token"
)

// contextKey は context.WithValue のキーに使う専用型。
// なぜ string ではなく専用型を使うか:
//   context.WithValue は interface{} をキーに取るため、異なるパッケージが
//   同じ文字列をキーに使うと衝突する。専用型を定義することで衝突を防ぐ。
//   これは Go の context パッケージのベストプラクティス。
type contextKey string

const claimsKey contextKey = "claims"

// Auth は Access Token を検証する認証ミドルウェア。
type Auth struct {
	verifier *token.Verifier
}

// NewAuth は新しい Auth ミドルウェアを作成する。
func NewAuth(verifier *token.Verifier) *Auth {
	return &Auth{verifier: verifier}
}

// Wrap は http.Handler をラップして認証チェックを追加する。
//
// 処理の流れ:
//   1. Authorization ヘッダーから Bearer トークンを抽出
//   2. トークンの署名とクレームを検証
//   3. 検証成功 → claims を context に格納して次のハンドラーへ
//   4. 検証失敗 → 401 Unauthorized を返却
func (a *Auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization ヘッダーの取得と解析
		// フォーマット: "Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "Authorizationヘッダーがありません")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "Authorizationヘッダーの形式が不正です（Bearer <token> が必要）")
			return
		}

		tokenString := parts[1]

		// トークンの検証
		claims, err := a.verifier.VerifyAccessToken(tokenString)
		if err != nil {
			// サーバーサイドには詳細なエラーをログ出力
			log.Printf("トークン検証失敗: %v", err)
			// クライアントにはセキュリティ上、詳細を返さない
			// なぜ: エラーの詳細が攻撃者にヒントを与える可能性がある
			writeError(w, http.StatusUnauthorized, "トークンが無効です")
			return
		}

		// 検証成功: claims を context に格納して次のハンドラーへ
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClaimsFromContext は context から Claims を取得する。
// 認証ミドルウェアを通過したハンドラーでのみ使用可能。
func ClaimsFromContext(ctx context.Context) *token.Claims {
	claims, _ := ctx.Value(claimsKey).(*token.Claims)
	return claims
}

// writeError はエラーレスポンスを JSON 形式で返す。
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
