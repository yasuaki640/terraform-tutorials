package middleware

import "net/http"

// CORS はローカル開発用の CORS ミドルウェアを返す。
//
// 学習ポイント: CORS（Cross-Origin Resource Sharing）は、
// ブラウザの Same-Origin Policy を緩和する仕組み。
// フロントエンド（localhost:3000）からバックエンド（localhost:8080）への
// リクエストを許可するために必要。
//
// 本番環境では:
// - Origin を厳密に制限する（ワイルドカード * は使わない）
// - 許可するメソッドとヘッダーを最小限にする
// - credentials を許可する場合は Access-Control-Allow-Origin に * を使えない
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 同一オリジンからのリクエスト（バックエンドが配信するHTML）も許可
			origin := r.Header.Get("Origin")
			if origin == allowedOrigin || origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Preflight リクエスト（OPTIONS）への応答
			// 学習ポイント: ブラウザは「単純でないリクエスト」の前に
			// OPTIONS リクエストを送信して、サーバーが CORS を許可しているか確認する。
			// Authorization ヘッダーを含むリクエストは「単純でない」ため、
			// Preflight が発生する。
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
