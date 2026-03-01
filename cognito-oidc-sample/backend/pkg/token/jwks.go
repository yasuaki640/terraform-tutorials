// Package token は JWT（JSON Web Token）の検証機能を提供する。
//
// 標準ライブラリのみで実装することで、go-oidc や golang-jwt ライブラリが
// 内部で行っている処理を理解する。
package token

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// JWKS（JSON Web Key Set）プロバイダー
// =============================================================================
// JWKSとは:
//   OIDC プロバイダー（Cognito）が公開する署名検証用の公開鍵セット。
//   JWT のヘッダーに含まれる kid（Key ID）で対応する公開鍵を特定し、
//   その公開鍵で JWT の署名を検証する。
//
// なぜ公開鍵で検証できるのか:
//   Cognito は秘密鍵で JWT に署名する（RSA-SHA256）。
//   対応する公開鍵は JWKS エンドポイントで公開されている。
//   公開鍵で署名を検証できれば、そのJWTは確かに Cognito が発行したものだと
//   証明できる（公開鍵暗号の原理）。
//
// なぜキャッシュが必要か:
//   リクエストごとに JWKS エンドポイントへ HTTP 通信すると遅延が発生する。
//   公開鍵はめったに変更されないため、TTL 付きでキャッシュするのが一般的。
//   go-oidc ライブラリも内部で同様のキャッシュを実装している。
// =============================================================================

// JWKSProvider は Cognito の JWKS エンドポイントから公開鍵を取得・キャッシュする。
type JWKSProvider struct {
	jwksURI   string
	keys      map[string]*rsa.PublicKey // kid -> RSA公開鍵
	mu        sync.RWMutex             // 読み取りと書き込みの排他制御
	lastFetch time.Time
	cacheTTL  time.Duration
}

// jwksResponse は JWKS エンドポイントの JSON レスポンスを表す。
// 仕様: https://tools.ietf.org/html/rfc7517
type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

// jwk は個々の JSON Web Key を表す。
// 各フィールドの意味:
//   - kid: Key ID。JWT ヘッダーの kid と照合して使用する公開鍵を特定する
//   - kty: Key Type。"RSA" = RSA暗号鍵（Cognito は RSA を使用）
//   - alg: Algorithm。"RS256" = RSA + SHA-256 署名
//   - use: 用途。"sig" = 署名（signature）用
//   - n:   RSA の modulus（法）。Base64URL エンコードされた巨大整数
//   - e:   RSA の exponent（指数）。Base64URL エンコードされた整数（通常 65537）
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewJWKSProvider は新しい JWKSProvider を作成する。
func NewJWKSProvider(jwksURI string) *JWKSProvider {
	return &JWKSProvider{
		jwksURI:  jwksURI,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: 1 * time.Hour,
	}
}

// GetKey は指定された kid に対応する RSA 公開鍵を返す。
// キャッシュが有効ならキャッシュから、期限切れなら JWKS エンドポイントから再取得する。
func (p *JWKSProvider) GetKey(kid string) (*rsa.PublicKey, error) {
	// まずリードロックでキャッシュを確認（複数goroutineが同時に読み取り可能）
	p.mu.RLock()
	if time.Since(p.lastFetch) < p.cacheTTL {
		if key, ok := p.keys[kid]; ok {
			p.mu.RUnlock()
			return key, nil
		}
	}
	p.mu.RUnlock()

	// キャッシュにない or 期限切れ → ライトロックで再取得
	p.mu.Lock()
	defer p.mu.Unlock()

	// ダブルチェック: 別の goroutine が先に取得済みかもしれない
	if time.Since(p.lastFetch) < p.cacheTTL {
		if key, ok := p.keys[kid]; ok {
			return key, nil
		}
	}

	if err := p.fetchKeys(); err != nil {
		return nil, fmt.Errorf("JWKSの取得に失敗: %w", err)
	}

	key, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid=%q に対応する公開鍵が見つかりません", kid)
	}
	return key, nil
}

// fetchKeys は JWKS エンドポイントから公開鍵を取得してキャッシュに格納する。
// 呼び出し元で排他ロックを取得済みであること。
func (p *JWKSProvider) fetchKeys() error {
	resp, err := http.Get(p.jwksURI)
	if err != nil {
		return fmt.Errorf("HTTPリクエスト失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKSエンドポイントが %d を返しました", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("JSONパース失敗: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		// RSA署名用の鍵のみを処理
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}

		pubKey, err := parseRSAPublicKey(k)
		if err != nil {
			return fmt.Errorf("kid=%q の公開鍵パースに失敗: %w", k.Kid, err)
		}
		newKeys[k.Kid] = pubKey
	}

	p.keys = newKeys
	p.lastFetch = time.Now()
	return nil
}

// parseRSAPublicKey は JWK から RSA 公開鍵を構築する。
//
// RSA公開鍵の構成要素:
//   - N（modulus）: 2つの巨大素数の積。公開鍵の本体。
//   - E（exponent）: 公開指数。通常 65537（0x10001）。
//
// JWK では N と E が Base64URL エンコードされた巨大整数として格納されている。
// これを Go の *rsa.PublicKey に変換する。
//
// Base64URL と標準 Base64 の違い:
//   標準 Base64:  + / =（パディング）
//   Base64URL:    - _ （パディングなし）
//   URL やファイル名に安全に使えるよう、記号を変換している（RFC 4648）。
func parseRSAPublicKey(k jwk) (*rsa.PublicKey, error) {
	// N（modulus）をデコード
	nBytes, err := base64URLDecode(k.N)
	if err != nil {
		return nil, fmt.Errorf("modulusのデコードに失敗: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// E（exponent）をデコード
	eBytes, err := base64URLDecode(k.E)
	if err != nil {
		return nil, fmt.Errorf("exponentのデコードに失敗: %w", err)
	}
	// exponent を int に変換（通常は小さい値なので int で十分）
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// base64URLDecode は Base64URL エンコードされた文字列をデコードする。
// Cognito の JWKS レスポンスにはパディング（=）が含まれない場合があるため、
// RawURLEncoding（パディングなし）を使用する。
// ただし、パディングが含まれている場合にも対応するため、事前に除去する。
func base64URLDecode(s string) ([]byte, error) {
	// パディングが含まれている場合に備えて除去
	s = strings.TrimRight(s, "=")
	return base64.RawURLEncoding.DecodeString(s)
}
