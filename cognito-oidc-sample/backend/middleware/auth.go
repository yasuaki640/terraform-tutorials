package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/yasuaki640/cognito-oidc-sample/backend/pkg/token"
)

// contextKey は context にセットする値のキー型。
// string ではなく独自型を使うことで、他パッケージとのキー衝突を防ぐ。
type contextKey string

const claimsKey contextKey = "claims"

// TokenVerifier はトークン検証のインターフェース
type TokenVerifier interface {
	VerifyAccessToken(tokenString string) (*token.Claims, error)
}

// Auth は Access Token を検証するミドルウェアを返す。
//
// 学習ポイント: ミドルウェアパターンは Go の HTTP サーバーで広く使われる。
// 認証・認可のロジックをハンドラーから分離することで、
// 各ハンドラーは認証済みであることを前提にビジネスロジックに集中できる。
//
// このパターンは「関心の分離」（Separation of Concerns）の実践。
// 認証 → 認可 → ビジネスロジック の順にミドルウェアチェーンを構成するのが一般的。
func Auth(verifier TokenVerifier) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Authorization ヘッダーから Bearer トークンを取得
			//    RFC 6750: Bearer トークンの利用方法
			//    形式: "Authorization: Bearer <token>"
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("Authorization ヘッダーがありません")
				http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				log.Printf("不正な Authorization ヘッダー形式: %s", authHeader[:min(len(authHeader), 20)])
				http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// 2. Access Token を検証
			//    学習ポイント: ここで以下の検証が行われる:
			//    - JWT の署名検証（JWKS 公開鍵で RS256 署名を検証）
			//    - iss（Issuer）の検証
			//    - token_use が "access" であることの検証
			//    - exp（有効期限）の検証
			claims, err := verifier.VerifyAccessToken(tokenString)
			if err != nil {
				log.Printf("トークン検証失敗: %v", err)
				http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// 3. Claims を context にセットして次のハンドラーへ
			//    学習ポイント: Go の context.Context を使ってリクエストスコープの
			//    データを伝播するのは、標準的なパターン。
			//    ただし context の使いすぎは結合度を高めるので注意。
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

// GetClaims は context から Claims を取得する。
func GetClaims(ctx context.Context) *token.Claims {
	claims, ok := ctx.Value(claimsKey).(*token.Claims)
	if !ok {
		return nil
	}
	return claims
}
