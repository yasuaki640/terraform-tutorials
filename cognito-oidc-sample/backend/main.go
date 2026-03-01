package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yasuaki640/cognito-oidc-sample/backend/handler"
	"github.com/yasuaki640/cognito-oidc-sample/backend/middleware"
	"github.com/yasuaki640/cognito-oidc-sample/backend/pkg/token"
)

// Config はアプリケーション設定を保持する。
// 環境変数から読み込む。
type Config struct {
	Port         string
	CognitoDomain string
	ClientID     string
	CallbackURL  string
	FrontendURL  string
	JwksURI      string
	Issuer       string
}

func loadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		CognitoDomain: mustGetEnv("COGNITO_DOMAIN"),
		ClientID:     mustGetEnv("COGNITO_CLIENT_ID"),
		CallbackURL:  getEnv("CALLBACK_URL", "http://localhost:8080/callback"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
		JwksURI:      mustGetEnv("COGNITO_JWKS_URI"),
		Issuer:       mustGetEnv("COGNITO_ISSUER"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("環境変数 %s が設定されていません", key)
	}
	return v
}

func main() {
	cfg := loadConfig()

	// JWKS キャッシュを初期化
	// 学習ポイント: JWKS（JSON Web Key Set）は JWT の署名検証に使う公開鍵セット。
	// Cognito の JWKS エンドポイントから取得し、キャッシュしておくことで
	// リクエストごとにネットワーク通信を発生させない。
	jwksCache := token.NewJWKSCache(cfg.JwksURI)

	// トークン検証器を初期化
	verifier := token.NewVerifier(jwksCache, cfg.Issuer, cfg.ClientID)

	// ハンドラーを初期化
	callbackHandler := handler.NewCallbackHandler(cfg.CognitoDomain, cfg.ClientID, cfg.CallbackURL)
	meHandler := handler.NewMeHandler()
	logoutHandler := handler.NewLogoutHandler(cfg.CognitoDomain, cfg.ClientID, cfg.FrontendURL)

	// CORS ミドルウェア
	// 学習ポイント: ブラウザの Same-Origin Policy により、
	// フロントエンド（localhost:3000）からバックエンド（localhost:8080）への
	// リクエストはデフォルトでブロックされる。CORS ヘッダーで許可する。
	corsMiddleware := middleware.CORS(cfg.FrontendURL)

	mux := http.NewServeMux()

	// GET / → フロントエンドの index.html を配信
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	// GET /config → フロントエンドに設定値を返す（シークレットは含めない）
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"cognito_domain":%q,"client_id":%q,"callback_url":%q,"frontend_url":%q}`,
			cfg.CognitoDomain, cfg.ClientID, cfg.CallbackURL, cfg.FrontendURL)
	})

	// POST /callback → 認可コード→トークン交換
	mux.HandleFunc("/callback", callbackHandler.Handle)

	// GET /me → ユーザー情報返却（要認証）
	// 学習ポイント: auth ミドルウェアが Access Token を検証し、
	// 検証成功時のみ meHandler に到達する。これが「認可」の実装パターン。
	authMiddleware := middleware.Auth(verifier)
	mux.HandleFunc("/me", authMiddleware(meHandler.Handle))

	// POST /logout → Cognito ログアウト URL を返却
	mux.HandleFunc("/logout", logoutHandler.Handle)

	// サーバー起動
	addr := ":" + cfg.Port
	log.Printf("サーバー起動: http://localhost%s", addr)
	log.Printf("フロントエンド: http://localhost%s/", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
