# =============================================================================
# マルチステージビルド
# =============================================================================
# なぜマルチステージビルドを使うか:
#   1. ビルドツール（Go コンパイラ、400MB+）を最終イメージに含めない
#   2. 最終イメージサイズを大幅に削減（数百MB → 数十MB）
#   3. 攻撃対象面（attack surface）の縮小 → セキュリティ向上
#   実務でも Docker イメージの最適化は基本中の基本。
# =============================================================================

# --- ビルドステージ ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Go モジュールファイルをコピー
# このプロジェクトは標準ライブラリのみ使用するため、go.sum は存在しない
COPY go.mod ./
COPY . .

# 静的バイナリをビルド
# CGO_ENABLED=0: C ライブラリに依存しない静的リンクバイナリを生成
#   なぜ: Alpine の musl libc ではなく、完全に独立したバイナリにする
#   これにより、scratch（空）イメージでも実行可能になる
# GOOS=linux: Linux 向けにクロスコンパイル
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend .

# --- 実行ステージ ---
# なぜ alpine を使うか:
#   scratch（空イメージ）よりデバッグが容易（sh, wget 等が使える）
#   かつ、ubuntu 等のフルイメージより遥かに小さい（5MB）
FROM alpine:3.19

# CA証明書をインストール
# なぜ: Cognito の HTTPS エンドポイントにアクセスするために必要
# Alpine のベースイメージには CA 証明書が含まれていない
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /backend .

EXPOSE 8080
CMD ["./backend"]
