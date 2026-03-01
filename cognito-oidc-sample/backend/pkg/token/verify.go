package token

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// JWT（JSON Web Token）手動検証
// =============================================================================
// JWT の構造（RFC 7519）:
//   Header.Payload.Signature（ドットで3つのパートに分かれる）
//
//   Header:    {"alg":"RS256","kid":"xxxxx"}    ← 署名アルゴリズムとキーID
//   Payload:   {"sub":"user-id","email":"..."}  ← クレーム（ユーザー情報等）
//   Signature: RSASHA256(base64url(header) + "." + base64url(payload), 秘密鍵)
//
// 検証手順:
//   1. JWT を "." で分割し、3パートあることを確認
//   2. Header をデコードして kid と alg を取得
//   3. alg が RS256 であることを確認（他のアルゴリズムは拒否）
//   4. kid で JWKS から公開鍵を取得
//   5. Header.Payload 部分の SHA-256 ハッシュを計算
//   6. 公開鍵で署名を検証（crypto/rsa.VerifyPKCS1v15）
//   7. Payload をデコードしてクレームを検証（iss, exp, token_use 等）
//
// なぜ手動で実装するか:
//   golang-jwt ライブラリの jwt.Parse() が内部で行っている処理を理解するため。
//   特に、既存の aws-cognito-sample では jwt.ParseUnverified() を使っており、
//   署名検証をスキップしている。これは本番環境では絶対に避けるべきプラクティス
//   （攻撃者が任意のクレームを持つ JWT を作成できてしまう）。
// =============================================================================

// 検証エラー定義
// 具体的なエラー型を定義することで、呼び出し元がエラーの種類を判別できる。
var (
	ErrMalformedToken   = errors.New("トークンの形式が不正です（Header.Payload.Signature の3パートが必要）")
	ErrUnsupportedAlg   = errors.New("サポートされていないアルゴリズムです（RS256のみ対応）")
	ErrKeyNotFound      = errors.New("JWKSに一致するキーが見つかりません")
	ErrInvalidSignature = errors.New("トークンの署名が無効です")
	ErrTokenExpired     = errors.New("トークンの有効期限が切れています")
	ErrInvalidIssuer    = errors.New("トークンの発行者（issuer）が一致しません")
	ErrInvalidAudience  = errors.New("トークンの対象者（audience/client_id）が一致しません")
	ErrInvalidTokenUse  = errors.New("トークンの用途（token_use）が不正です")
)

// Verifier は JWT の署名検証とクレーム検証を行う。
type Verifier struct {
	jwks     *JWKSProvider
	issuer   string
	clientID string
}

// NewVerifier は新しい Verifier を作成する。
func NewVerifier(jwks *JWKSProvider, issuer, clientID string) *Verifier {
	return &Verifier{
		jwks:     jwks,
		issuer:   issuer,
		clientID: clientID,
	}
}

// jwtHeader は JWT ヘッダーの構造体。
type jwtHeader struct {
	Alg string `json:"alg"` // 署名アルゴリズム（RS256）
	Kid string `json:"kid"` // JWKS内の公開鍵を特定するキーID
}

// Claims は JWT のペイロード（クレーム）を表す。
//
// Cognito 固有の注意点:
//   - Access Token: aud クレームを持たず、代わりに client_id を使用
//   - ID Token: aud クレームにクライアントIDが入る
//   この違いは Cognito の仕様であり、検証ロジックで考慮する必要がある。
type Claims struct {
	Sub       string `json:"sub"`        // Subject: ユーザーの一意識別子
	Email     string `json:"email"`      // メールアドレス
	Iss       string `json:"iss"`        // Issuer: トークン発行者
	Aud       string `json:"aud"`        // Audience: ID Token のみ
	ClientID  string `json:"client_id"`  // Access Token のクライアントID
	TokenUse  string `json:"token_use"`  // "access" or "id"
	Exp       int64  `json:"exp"`        // Expiration: 有効期限（Unix timestamp）
	Iat       int64  `json:"iat"`        // Issued At: 発行時刻
	AuthTime  int64  `json:"auth_time"`  // 認証が行われた時刻
	Username  string `json:"username"`   // ユーザー名（Cognito独自）
	Scope     string `json:"scope"`      // スコープ（Access Token のみ）
}

// VerifyAccessToken は Access Token を検証し、クレームを返す。
//
// Access Token の検証で確認する項目:
//   1. 署名の正当性（RSA-SHA256）
//   2. issuer が期待値と一致するか
//   3. token_use が "access" であるか
//   4. client_id が期待値と一致するか（Access Token は aud ではなく client_id を使用）
//   5. 有効期限が切れていないか
func (v *Verifier) VerifyAccessToken(tokenString string) (*Claims, error) {
	claims, err := v.verifySignatureAndDecode(tokenString)
	if err != nil {
		return nil, err
	}

	// issuer の検証
	if claims.Iss != v.issuer {
		return nil, fmt.Errorf("%w: got=%q, want=%q", ErrInvalidIssuer, claims.Iss, v.issuer)
	}

	// token_use の検証
	if claims.TokenUse != "access" {
		return nil, fmt.Errorf("%w: got=%q, want=%q", ErrInvalidTokenUse, claims.TokenUse, "access")
	}

	// client_id の検証（Access Token 固有: aud ではなく client_id クレームを使用）
	if claims.ClientID != v.clientID {
		return nil, fmt.Errorf("%w: got=%q, want=%q", ErrInvalidAudience, claims.ClientID, v.clientID)
	}

	// 有効期限の検証
	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// verifySignatureAndDecode は JWT の署名を検証し、ペイロードをデコードする。
// Access Token と ID Token で共通の処理。
func (v *Verifier) verifySignatureAndDecode(tokenString string) (*Claims, error) {
	// ステップ1: JWT を "." で3パートに分割
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrMalformedToken
	}
	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// ステップ2: ヘッダーをデコードして kid と alg を取得
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return nil, fmt.Errorf("ヘッダーのBase64URLデコードに失敗: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("ヘッダーのJSONパースに失敗: %w", err)
	}

	// ステップ3: アルゴリズムの検証
	// ★ セキュリティ上の重要ポイント ★
	// JWT の alg フィールドを無条件に信用してはいけない。
	// 攻撃者が alg を "none" や "HS256" に差し替えて署名検証を回避する
	// 「アルゴリズム混同攻撃」が知られている。
	// 対策: サーバー側で許可するアルゴリズムを固定し、それ以外は拒否する。
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("%w: got=%q", ErrUnsupportedAlg, header.Alg)
	}

	// ステップ4: kid で JWKS から公開鍵を取得
	pubKey, err := v.jwks.GetKey(header.Kid)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyNotFound, err)
	}

	// ステップ5: 署名対象データ（Header.Payload）の SHA-256 ハッシュを計算
	// JWT の署名は "base64url(header).base64url(payload)" に対して行われる。
	// デコード前の Base64URL 文字列をそのまま使うことが重要。
	signedData := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signedData))

	// ステップ6: 署名をデコードして RSA 署名検証
	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("署名のBase64URLデコードに失敗: %w", err)
	}

	// RSA-PKCS1v15 署名検証
	// この関数は、公開鍵で署名を「復号」し、結果がハッシュ値と一致するか確認する。
	// 一致すれば、対応する秘密鍵（Cognito が保持）で署名されたことが証明される。
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, ErrInvalidSignature
	}

	// ステップ7: ペイロードをデコードしてクレームを取得
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("ペイロードのBase64URLデコードに失敗: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("ペイロードのJSONパースに失敗: %w", err)
	}

	return &claims, nil
}
