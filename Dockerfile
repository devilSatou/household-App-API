# ビルドステージ（Debianベース）
FROM golang:1.21-bookworm AS builder

# SQLite用のビルド環境
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# CGO有効でビルド（SQLite対応）
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# 実行ステージ（軽量Debianベース）
FROM debian:bookworm-slim

# 実行時に必要なライブラリ
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

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
