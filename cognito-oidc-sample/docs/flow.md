# OIDC Authorization Code Flow + PKCE フロー図解

## 全体フロー

```
┌──────────┐     ┌──────────┐     ┌──────────────┐
│  Browser │     │  Backend │     │   Cognito    │
│(Frontend)│     │ (Go API) │     │ (認可サーバー) │
└────┬─────┘     └────┬─────┘     └──────┬───────┘
     │                │                   │
     │ 1. ログインボタンクリック            │
     │ ─ code_verifier 生成               │
     │ ─ code_challenge = SHA256(verifier) │
     │ ─ state 生成（CSRF対策）             │
     │                │                   │
     │ 2. /authorize にリダイレクト ──────────────────►│
     │    ?response_type=code             │
     │    &client_id=xxx                  │
     │    &redirect_uri=xxx               │
     │    &scope=openid email profile     │
     │    &code_challenge=xxx             │
     │    &code_challenge_method=S256     │
     │    &state=xxx                      │
     │                │                   │
     │                │    3. Hosted UI 表示
     │◄───────────────────────────────────│
     │                │                   │
     │ 4. ユーザーがID/PW入力              │
     │────────────────────────────────────►│
     │                │                   │
     │ 5. 認可コード付きでリダイレクト       │
     │◄───────────────────────────────────│
     │    ?code=AUTH_CODE&state=xxx        │
     │                │                   │
     │ 6. state 検証（CSRF チェック）        │
     │                │                   │
     │ 7. POST /callback ──────►│          │
     │    {code, code_verifier} │          │
     │                │                   │
     │                │ 8. POST /oauth2/token ────►│
     │                │    grant_type=authorization_code
     │                │    code=AUTH_CODE  │
     │                │    code_verifier=xxx
     │                │                   │
     │                │    9. トークン返却  │
     │                │◄──────────────────│
     │                │    {id_token,      │
     │                │     access_token,  │
     │                │     refresh_token} │
     │                │                   │
     │ 10. トークン返却│                   │
     │◄───────────────│                   │
     │                │                   │
     │ 11. GET /me ──────────►│            │
     │    Authorization: Bearer access_token
     │                │                   │
     │                │ 12. JWT署名検証     │
     │                │ ─ JWKS公開鍵取得 ─────────►│
     │                │◄──────────────────│
     │                │ ─ RS256署名検証    │
     │                │ ─ claims検証       │
     │                │   (iss, exp, token_use)
     │                │                   │
     │ 13. ユーザー情報│                   │
     │◄───────────────│                   │
     │    {sub, email} │                   │
     │                │                   │
```

## 各トークンの役割

| トークン | 用途 | 形式 | 有効期限 |
|---------|------|------|---------|
| ID Token | ユーザーの「認証」情報 | JWT (署名付き) | 1時間 |
| Access Token | API アクセスの「認可」 | JWT (署名付き) | 1時間 |
| Refresh Token | 新しいトークン取得 | Opaque (JWT ではない) | 30日 |

## PKCE の役割

```
攻撃シナリオ（PKCE なし）:
1. 攻撃者がリダイレクト URI を傍受
2. 認可コードを横取り
3. トークン交換に成功 → アカウント乗っ取り

PKCE での防御:
1. 攻撃者がリダイレクト URI を傍受
2. 認可コードを横取り
3. トークン交換に code_verifier が必要
4. code_verifier は攻撃者が知り得ない → トークン交換失敗
```

## JWT 署名検証の流れ

```
JWT: xxxxx.yyyyy.zzzzz
      │       │      │
      │       │      └── Signature (Base64URL)
      │       └── Payload (Base64URL) ── デコード → Claims (JSON)
      └── Header (Base64URL) ── デコード → {kid, alg}

検証手順:
1. Header から kid を取得
2. JWKS エンドポイントから kid に対応する公開鍵を取得
3. SHA256(Header.Payload) を計算
4. 公開鍵で Signature を検証（RSA PKCS#1 v1.5）
5. Payload の claims を検証（iss, aud, exp, token_use）
```
