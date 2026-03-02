# OIDC Authorization Code + PKCE フロー詳解

## 全体フロー（シーケンス図）

```
ブラウザ (SPA)                    バックエンド (Go)                Cognito
     |                                  |                            |
     |  1. ログインボタンクリック         |                            |
     |  code_verifier = random()        |                            |
     |  code_challenge = SHA256(code_verifier)                       |
     |  state = random()                |                            |
     |  sessionStorage に保存           |                            |
     |                                  |                            |
     |  2. /authorize にリダイレクト -------------------------------->|
     |     response_type=code           |                            |
     |     client_id=xxx                |                            |
     |     redirect_uri=http://...      |                            |
     |     scope=openid email profile   |                            |
     |     code_challenge=xxx           |                            |
     |     code_challenge_method=S256   |                            |
     |     state=xxx                    |                            |
     |                                  |                            |
     |                                  |  3. Hosted UI でログイン画面|
     |  <-------------------------------------------------------------|
     |                                  |                            |
     |  4. ユーザーが認証情報を入力      |                            |
     |  ------------------------------------------------------------->|
     |                                  |                            |
     |  5. コールバック（リダイレクト）    |                            |
     |  <-------------------------------------------------------------|
     |     /callback?code=AUTH_CODE&state=xxx                        |
     |                                  |                            |
     |  6. state を検証（CSRF チェック）  |                            |
     |  sessionStorage の値と照合       |                            |
     |                                  |                            |
     |  7. POST /callback ------------->|                            |
     |     { code, code_verifier }      |                            |
     |                                  |  8. POST /oauth2/token --->|
     |                                  |     grant_type=             |
     |                                  |       authorization_code   |
     |                                  |     code=AUTH_CODE         |
     |                                  |     code_verifier=xxx      |
     |                                  |     client_id=xxx          |
     |                                  |     redirect_uri=xxx       |
     |                                  |                            |
     |                                  |  9. Cognito が検証:        |
     |                                  |     SHA256(code_verifier)  |
     |                                  |       == code_challenge ?  |
     |                                  |                            |
     |                                  |  10. トークンレスポンス <---|
     |                                  |     { access_token,        |
     |  11. トークンを受信 <------------|       id_token,            |
     |     メモリ変数に保存             |       refresh_token }      |
     |                                  |                            |
     |  12. GET /me ------------------->|                            |
     |     Authorization: Bearer xxx    |                            |
     |                                  |  13. JWT 署名検証          |
     |                                  |     a. JWKS から公開鍵取得  |
     |                                  |        (キャッシュ付き)  -->|
     |                                  |     b. RSA 署名検証        |<-- .well-known/jwks.json
     |                                  |     c. claims 検証         |
     |                                  |        (iss, exp, etc.)    |
     |  14. ユーザー情報 <--------------|                            |
     |     { sub, email, username }     |                            |
     |                                  |                            |
     |  15. ユーザー情報を表示          |                            |
     |                                  |                            |
```

## 各ステップの詳細

### ステップ 1-2: 認可リクエスト（Authorization Request）

PKCE の核心部分。ブラウザが以下を生成:

- **code_verifier**: 43〜128文字の暗号学的ランダム文字列
- **code_challenge**: `BASE64URL(SHA256(code_verifier))`
- **state**: CSRF 対策用のランダム文字列

`code_verifier` は `sessionStorage` に保存し、Cognito には `code_challenge` のみを送る。

### ステップ 3-5: ユーザー認証とコールバック

Cognito Hosted UI でユーザーが認証すると、設定済みの `redirect_uri` に認可コードが付与されてリダイレクトされる。

```
http://localhost:8080/callback?code=AUTH_CODE&state=STATE_VALUE
```

### ステップ 6: state の検証

コールバックの `state` パラメーターを `sessionStorage` の値と照合。
一致しなければ CSRF 攻撃の可能性があるため処理を中断。

### ステップ 7-10: トークン交換

バックエンドが Cognito の `/oauth2/token` エンドポイントに `code` と `code_verifier` を送信。
Cognito は `SHA256(code_verifier) == code_challenge` を検証し、一致すればトークンを発行。

### ステップ 12-14: JWT 署名検証

バックエンドが Access Token の RSA-SHA256 署名を JWKS の公開鍵で検証。

## PKCE がない場合の脆弱性

```
攻撃者                  被害者のブラウザ              Cognito
  |                          |                         |
  |  認可コードを傍受         |                         |
  |  (リダイレクトURLの       |  /authorize ---------->|
  |   スキーム乗っ取り等)     |  <--------- code ------|
  |  <---- code を横取り ----|                         |
  |                          |                         |
  |  /oauth2/token ---------------------------------------->|
  |     code=横取りしたコード  |                         |
  |     （code_verifier 不要） |                         |
  |  <------- access_token --------------------------------|
  |                          |                         |
  |  被害者としてアクセス可能！ |                         |
```

PKCE があれば、攻撃者は `code_verifier` を持たないため、
横取りした認可コードでトークンを取得できない。

## 認可フロー比較表

| 項目 | Implicit Flow | Authorization Code | Auth Code + PKCE |
|------|--------------|-------------------|-----------------|
| トークンの受け渡し | URL フラグメント | バックエンド経由 | バックエンド経由 |
| クライアントシークレット | 不要 | 必要 | 不要 |
| リフレッシュトークン | なし | あり | あり |
| セキュリティ | 低（非推奨） | 高 | 高 |
| 用途 | ─ | サーバーサイドアプリ | SPA / モバイルアプリ |
| RFC | ─ | RFC 6749 | RFC 7636 |

## トークンの種類

| トークン | 用途 | 有効期限 | 保存場所 |
|---------|------|---------|---------|
| Access Token | API 認可（Bearer Token） | 1時間 | メモリ変数 |
| ID Token | ユーザー情報の取得 | 1時間 | メモリ変数 |
| Refresh Token | トークンの更新 | 30日 | メモリ変数 |

### Access Token vs ID Token

- **Access Token**: API へのアクセス権を表す。`token_use: "access"`。Cognito 固有: `aud` クレームがなく `client_id` を使用。
- **ID Token**: ユーザーの身元情報を表す。`token_use: "id"`。`aud` クレームにクライアント ID が入る。

## JWT 署名検証のフロー

```
JWT: eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQifQ.eyJzdWIiOiJ1c2VyMSJ9.SIGNATURE
      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^
              Header (Base64URL)                Payload (Base64URL)     Signature

1. "." で分割 → [Header, Payload, Signature]

2. Header をデコード → { "alg": "RS256", "kid": "1234" }

3. alg == "RS256" を確認（他のアルゴリズムは拒否）

4. kid = "1234" で JWKS から公開鍵を取得

5. hash = SHA256("eyJhbGci...ImtpZCI6IjEyMzQifQ.eyJzdWIiOiJ1c2VyMSJ9")
   （Header.Payload の Base64URL 文字列を結合してハッシュ）

6. rsa.VerifyPKCS1v15(公開鍵, SHA256, hash, Signature)
   → 成功: Cognito が発行した正当なトークン
   → 失敗: 改ざんまたは不正なトークン

7. Payload をデコード → Claims を検証
   - iss == "https://cognito-idp.{region}.amazonaws.com/{pool_id}"
   - token_use == "access"
   - client_id == 期待するクライアントID
   - exp > 現在時刻
```
