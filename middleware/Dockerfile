# Build Stage
FROM golang:1.25 AS builder
WORKDIR /app

# Go モジュールファイルのコピーと依存関係の取得
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .
RUN go mod tidy

# テストの実行 (テスト失敗時はビルドが中断されます)
# 仮想ミドルウェア(middleware)に関連するテストのみを実行します
RUN go test -v ./middleware/...

# サーバーのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server ./middleware/cmd/server

# Final Stage
FROM debian:stable-slim

COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]