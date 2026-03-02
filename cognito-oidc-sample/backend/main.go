// =============================================================================
// cognito-oidc-sample バックエンドサーバー
// =============================================================================
// このアプリケーションは OIDC Authorization Code + PKCE フローを
// 標準ライブラリのみで実装した学習用サンプル。
//
// 外部ライブラリを一切使わない理由:
//   - go-oidc, oauth2 ライブラリの内部動作を理解するため
//   - gorilla/mux → Go 1.22 の標準 ServeMux（メソッドベースルーティング対応）
//   - gorilla/sessions → トークンベース認証（サーバーサイドセッション不要）
//   - godotenv → Docker Compose の env_file で代替
//
// アーキテクチャ:
//   - 同一オリジン配信: フロントエンド（静的HTML）とAPI を同じサーバーで提供
//   - CORS 不要: フロントエンドとバックエンドが同じオリジン（localhost:8080）
//   - 依存注入: グローバル変数を避け、ハンドラーに必要な依存を明示的に渡す
// =============================================================================
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yasuaki640/terraform-tutorials/cognito-oidc-sample/backend/handler"
	"github.com/yasuaki640/terraform-tutorials/cognito-oidc-sample/backend/middleware"
	"github.com/yasuaki640/terraform-tutorials/cognito-oidc-sample/backend/pkg/token"
)

// config はアプリケーション設定を保持する構造体。
// 環境変数から読み込む（12 Factor App の原則: 設定をコードから分離）。
type config struct {
	CognitoDomain     string
	ClientID          string
	RedirectURI       string
	LogoutRedirectURI string
	JWKSURI           string
	Issuer            string
}

func main() {
	// 環境変数から設定を読み込む
	cfg := mustLoadConfig()

	// JWKSプロバイダーを初期化（公開鍵をキャッシュ付きで取得）
	jwksProvider := token.NewJWKSProvider(cfg.JWKSURI)

	// トークン検証器を初期化
	verifier := token.NewVerifier(jwksProvider, cfg.Issuer, cfg.ClientID)

	// 認証ミドルウェアを作成
	authMiddleware := middleware.NewAuth(verifier)

	// ハンドラーを初期化（依存注入）
	configHandler := handler.NewConfig(cfg.CognitoDomain, cfg.ClientID, cfg.RedirectURI)
	callbackHandler := handler.NewCallback(cfg.CognitoDomain, cfg.ClientID, cfg.RedirectURI)
	meHandler := handler.NewMe()
	logoutHandler := handler.NewLogout(cfg.CognitoDomain, cfg.ClientID, cfg.LogoutRedirectURI)

	// ルーティング設定
	// Go 1.22 以降の http.NewServeMux はメソッドベースルーティングに対応。
	// "GET /me" のように書くことで、gorilla/mux なしでメソッド+パスのルーティングが可能。
	mux := http.NewServeMux()

	// フロントエンド配信（静的ファイル）
	// http.FileServer は "/" で index.html を自動的に配信する。
	frontendFS := http.FileServer(http.Dir("./frontend"))
	mux.Handle("GET /", frontendFS)

	// Cognito からのリダイレクト先（GET /callback）
	// Cognito Hosted UI は認証後に callback URL にリダイレクトする。
	// フロントエンドの index.html を配信し、JS がクエリパラメーターを処理する。
	// 注意: POST /callback（トークン交換API）とは HTTP メソッドで区別される。
	mux.HandleFunc("GET /callback", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/index.html")
	})

	// API エンドポイント
	mux.HandleFunc("GET /config", configHandler.Handle)
	mux.HandleFunc("POST /callback", callbackHandler.Handle)
	mux.Handle("GET /me", authMiddleware.Wrap(http.HandlerFunc(meHandler.Handle)))
	mux.HandleFunc("POST /logout", logoutHandler.Handle)

	addr := ":8080"
	fmt.Printf("サーバー起動: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// mustLoadConfig は環境変数から設定を読み込む。
// 必須の環境変数が未設定の場合はプロセスを終了する。
func mustLoadConfig() config {
	return config{
		CognitoDomain:     mustEnv("COGNITO_DOMAIN"),
		ClientID:          mustEnv("CLIENT_ID"),
		RedirectURI:       mustEnv("REDIRECT_URI"),
		LogoutRedirectURI: mustEnv("LOGOUT_REDIRECT_URI"),
		JWKSURI:           mustEnv("JWKS_URI"),
		Issuer:            mustEnv("ISSUER"),
	}
}

// mustEnv は環境変数を取得する。未設定の場合はログを出力してプロセスを終了する。
// godotenv を使わない代わりに、Docker Compose の env_file で .env を読み込む。
// ローカル実行時は手動で export するか、シェルで .env を source する。
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("必須の環境変数 %s が設定されていません。.env ファイルを確認してください。", key)
	}
	return v
}
