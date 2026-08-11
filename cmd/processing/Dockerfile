# Build stage for Processing Unit (gRPC server)
FROM golang:1.25 AS builder
WORKDIR /app

# モジュールファイルと依存取得
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .

# テストの実行 (テスト失敗時はビルドが中断されます)
# 処理ユニット(processing)およびコア・共有ロジックに関連するテストのみを実行します
RUN go test -v ./cmd/processing/... ./internal/... ./pkg/... ./shared/...

WORKDIR /app/cmd/processing
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/processing .

# ランタイムイメージ
FROM debian:stable-slim

COPY --from=builder /app/processing /processing
EXPOSE 50051
ENTRYPOINT ["/processing"]