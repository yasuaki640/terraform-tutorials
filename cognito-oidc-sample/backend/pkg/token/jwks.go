package token

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JWKSCache は JWKS（JSON Web Key Set）の公開鍵をキャッシュする。
//
// 学習ポイント: JWKS は JWT の署名検証に使う公開鍵セットを配布するための仕様（RFC 7517）。
// Cognito は JWKS エンドポイントで RSA 公開鍵を公開しており、
// この鍵を使って JWT の署名が正当かを検証する。
//
// なぜキャッシュが必要か:
// - リクエストごとに JWKS エンドポイントにアクセスするのは非効率
// - Cognito の鍵ローテーションは頻繁ではない（数週間〜数ヶ月単位）
// - ただし、kid が見つからない場合は再取得を試みる必要がある
type JWKSCache struct {
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	mu        sync.RWMutex
	jwksURI   string
	ttl       time.Duration
}

// jwksResponse は JWKS エンドポイントのレスポンス
type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

// jwk は JSON Web Key の構造体
// 学習ポイント: JWK は以下のフィールドを持つ:
// - kty: 鍵の種類（RSA, EC 等）
// - kid: 鍵の識別子。JWT の Header にも含まれ、どの鍵で署名されたかを示す
// - use: 鍵の用途（sig = 署名, enc = 暗号化）
// - alg: アルゴリズム（RS256 等）
// - n: RSA 公開鍵のモジュラス（Base64URL エンコード）
// - e: RSA 公開鍵の指数（Base64URL エンコード）
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewJWKSCache(jwksURI string) *JWKSCache {
	return &JWKSCache{
		keys:    make(map[string]*rsa.PublicKey),
		jwksURI: jwksURI,
		ttl:     1 * time.Hour,
	}
}

// GetKey は kid に対応する RSA 公開鍵を返す。
// キャッシュが期限切れの場合、または kid が見つからない場合は再取得を試みる。
func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	// まずキャッシュから取得を試みる
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Since(c.fetchedAt) < c.ttl {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// キャッシュミスまたは期限切れの場合、JWKS を再取得
	// 学習ポイント: 鍵ローテーション時に新しい kid が来る可能性がある。
	// その場合はキャッシュを更新して再取得する。
	if err := c.refresh(); err != nil {
		return nil, fmt.Errorf("JWKS の取得に失敗: %w", err)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid=%s に対応する公開鍵が見つかりません", kid)
	}
	return key, nil
}

// refresh は JWKS エンドポイントから公開鍵を取得してキャッシュを更新する
func (c *JWKSCache) refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ダブルチェック: 他のゴルーチンが既に更新済みかもしれない
	if time.Since(c.fetchedAt) < 5*time.Second {
		return nil
	}

	log.Printf("JWKS を取得中: %s", c.jwksURI)

	resp, err := http.Get(c.jwksURI)
	if err != nil {
		return fmt.Errorf("JWKS エンドポイントへのリクエスト失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("JWKS レスポンスの読み取り失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS エンドポイントがエラーを返しました: status=%d", resp.StatusCode)
	}

	var jwksResp jwksResponse
	if err := json.Unmarshal(body, &jwksResp); err != nil {
		return fmt.Errorf("JWKS レスポンスのパース失敗: %w", err)
	}

	// 公開鍵をパースしてキャッシュに格納
	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwksResp.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}

		pubKey, err := parseRSAPublicKey(key)
		if err != nil {
			log.Printf("公開鍵のパース失敗 (kid=%s): %v", key.Kid, err)
			continue
		}
		newKeys[key.Kid] = pubKey
	}

	c.keys = newKeys
	c.fetchedAt = time.Now()
	log.Printf("JWKS キャッシュ更新完了: %d 個の公開鍵", len(newKeys))

	return nil
}

// parseRSAPublicKey は JWK から RSA 公開鍵を構築する。
//
// 学習ポイント: RSA 公開鍵は (n, e) の2つのパラメーターで構成される。
// - n（モジュラス）: 非常に大きな整数（2048bit 以上）
// - e（指数）: 通常 65537（0x10001）
// JWKS ではこれらが Base64URL エンコードされた big-endian バイト列として提供される。
func parseRSAPublicKey(key jwk) (*rsa.PublicKey, error) {
	// n（モジュラス）をデコード
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("n のデコード失敗: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// e（指数）をデコード
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("e のデコード失敗: %w", err)
	}
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
