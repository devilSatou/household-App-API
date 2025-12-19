# ビルドステージ
FROM golang:1.21-alpine AS builder

# SQLite用のビルド環境
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# CGO有効でビルド（SQLite対応）
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# 実行ステージ
FROM alpine:latest

# 実行時に必要なライブラリ
RUN apk add --no-cache ca-certificates tzdata

# タイムゾーン設定
ENV TZ=Asia/Tokyo

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /build/main .
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/static ./static

# データディレクトリ作成
RUN mkdir -p /app/data

# ポート公開
EXPOSE 8080

# 実行
CMD ["./main"]
