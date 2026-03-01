package token

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenUse はトークンの種別を表す
type TokenUse string

const (
	TokenUseID     TokenUse = "id"
	TokenUseAccess TokenUse = "access"
)

// Claims は JWT のペイロード（claims）を表す構造体。
//
// 学習ポイント: JWT の claims には「登録済み claims」と「カスタム claims」がある。
// 登録済み claims（RFC 7519 Section 4.1）:
// - iss (Issuer): トークン発行者
// - sub (Subject): トークンの主体（ユーザー識別子）
// - aud (Audience): トークンの受取人
// - exp (Expiration): 有効期限
// - iat (Issued At): 発行時刻
//
// Cognito 独自の claims:
// - token_use: "id" または "access"
// - cognito:username: Cognito ユーザー名
// - email: メールアドレス（スコープに email が含まれる場合）
type Claims struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Iss      string `json:"iss"`
	Aud      string `json:"aud"`
	TokenUse string `json:"token_use"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	ClientID string `json:"client_id"` // Access Token の場合に含まれる
}

// jwtHeader は JWT の Header 部分
type jwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

// Verifier は JWT の検証を行う
type Verifier struct {
	jwksCache *JWKSCache
	issuer    string
	clientID  string
}

func NewVerifier(jwksCache *JWKSCache, issuer, clientID string) *Verifier {
	return &Verifier{
		jwksCache: jwksCache,
		issuer:    issuer,
		clientID:  clientID,
	}
}

// VerifyIDToken は ID Token を検証して claims を返す
func (v *Verifier) VerifyIDToken(tokenString string) (*Claims, error) {
	return v.verifyToken(tokenString, TokenUseID)
}

// VerifyAccessToken は Access Token を検証して claims を返す
func (v *Verifier) VerifyAccessToken(tokenString string) (*Claims, error) {
	return v.verifyToken(tokenString, TokenUseAccess)
}

// verifyToken は JWT を検証して claims を返す。
//
// 学習ポイント: JWT の検証手順:
// 1. JWT を "." で3分割（Header.Payload.Signature）
// 2. Header をデコードして kid（Key ID）と alg（アルゴリズム）を取得
// 3. alg が期待するアルゴリズム（RS256）であることを確認
// 4. kid で JWKS キャッシュから公開鍵を取得
// 5. Header.Payload の SHA-256 ハッシュを計算
// 6. 公開鍵で Signature を検証（RSA PKCS#1 v1.5）
// 7. Payload をデコードして claims を検証（iss, aud, exp 等）
func (v *Verifier) verifyToken(tokenString string, expectedUse TokenUse) (*Claims, error) {
	// 1. JWT を "." で3分割
	//    学習ポイント: JWT は Header.Payload.Signature の3部構成。
	//    各部分は Base64URL エンコードされている。
	//    JWT は「暗号化」ではなく「署名」であることに注意。
	//    中身は誰でもデコードして読める。署名は「改ざんされていないこと」を保証する。
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("不正な JWT 形式: %d 部分に分割されました（期待値: 3）", len(parts))
	}

	headerPart := parts[0]
	payloadPart := parts[1]
	signaturePart := parts[2]

	// 2. Header をデコードして kid と alg を取得
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerPart)
	if err != nil {
		return nil, fmt.Errorf("JWT Header のデコード失敗: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("JWT Header のパース失敗: %w", err)
	}

	// 3. alg の検証
	//    学習ポイント: alg を検証しないと「alg: none」攻撃（署名なしの JWT を受け入れてしまう）や
	//    「RS256 → HS256 すり替え攻撃」（公開鍵を HMAC の秘密鍵として悪用）を受ける可能性がある。
	//    Cognito は RS256 を使用するので、それ以外は拒否する。
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("サポートされていないアルゴリズム: %s（期待値: RS256）", header.Alg)
	}

	// 4. JWKS キャッシュから公開鍵を取得
	publicKey, err := v.jwksCache.GetKey(header.Kid)
	if err != nil {
		return nil, fmt.Errorf("公開鍵の取得失敗: %w", err)
	}

	// 5. 署名対象データ（Header.Payload）の SHA-256 ハッシュを計算
	//    学習ポイント: 署名対象は Base64URL エンコードされた Header と Payload を
	//    "." で結合した文字列。デコード前のデータに対して署名検証する。
	signedData := headerPart + "." + payloadPart
	hash := sha256.Sum256([]byte(signedData))

	// 6. 署名をデコードして RSA 署名検証
	//    学習ポイント: RSA PKCS#1 v1.5 署名検証は:
	//    1. Signature を公開鍵で「復号」
	//    2. 復号結果が SignedData のハッシュと一致するか確認
	//    これにより、秘密鍵の保持者（= Cognito）が署名したことを数学的に証明する。
	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return nil, fmt.Errorf("署名のデコード失敗: %w", err)
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("JWT 署名検証失敗: %w", err)
	}

	// 7. Payload をデコードして claims を検証
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return nil, fmt.Errorf("JWT Payload のデコード失敗: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("JWT Claims のパース失敗: %w", err)
	}

	// Claims の検証
	if err := v.validateClaims(&claims, expectedUse); err != nil {
		return nil, err
	}

	return &claims, nil
}

// validateClaims は JWT の claims を検証する
func (v *Verifier) validateClaims(claims *Claims, expectedUse TokenUse) error {
	// iss（Issuer）の検証
	// 学習ポイント: iss が期待する発行者と一致しない場合、
	// 別の Cognito User Pool や別サービスが発行したトークンかもしれない。
	if claims.Iss != v.issuer {
		return fmt.Errorf("iss が一致しません: got=%s, want=%s", claims.Iss, v.issuer)
	}

	// token_use の検証
	// 学習ポイント: ID Token と Access Token で検証項目が異なる:
	// - ID Token: aud = client_id（このトークンは「誰か」を証明する）
	// - Access Token: aud は検証しない（Cognito の場合）、token_use = "access"
	// この違いが「認証」と「認可」の分離を体現している
	if claims.TokenUse != string(expectedUse) {
		return fmt.Errorf("token_use が一致しません: got=%s, want=%s", claims.TokenUse, expectedUse)
	}

	// aud（Audience）の検証（ID Token のみ）
	// 学習ポイント: Access Token では aud の代わりに client_id フィールドを使う（Cognito の仕様）
	if expectedUse == TokenUseID {
		if claims.Aud != v.clientID {
			return fmt.Errorf("aud が一致しません: got=%s, want=%s", claims.Aud, v.clientID)
		}
	}

	// exp（Expiration）の検証
	// 学習ポイント: exp は Unix タイムスタンプ（秒）。
	// 現在時刻より過去であればトークンは期限切れ。
	// 本番では数秒のクロックスキューを許容することもある。
	now := time.Now().Unix()
	if claims.Exp < now {
		return fmt.Errorf("トークンが期限切れです: exp=%d, now=%d", claims.Exp, now)
	}

	return nil
}
