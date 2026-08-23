# Build stage for Simulator (HTTP client)
FROM golang:1.25 AS builder
WORKDIR /app

# モジュールファイルと依存取得
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .

# テストの実行 (テスト失敗時はビルドが中断されます)
RUN go test -v ./cmd/simulator/...

WORKDIR /app/cmd/simulator
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/simulator .

# ランタイムイメージ
FROM debian:stable-slim

# SSL証明書をインストール（HTTPリクエスト送信に必要な場合があるため）
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/simulator /simulator
ENTRYPOINT ["/simulator"]
