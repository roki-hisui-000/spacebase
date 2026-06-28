#!/bin/bash

# エラー時にスクリプトを終了
set -e

# 各種コマンドのヘルプ表示
show_help() {
    echo "======================================================================"
    echo " Spacebase Development Skills Commands (skills.sh)"
    echo "======================================================================"
    echo " 開発やAIエージェントの操作をサポートするための便利なコマンド集です。"
    echo ""
    echo " 使い方: ./skills.sh <command>"
    echo ""
    echo " [コンテナ操作]"
    echo "   up               - 全てのコンテナをバックグラウンドで起動"
    echo "   down             - 全てのコンテナを停止・削除"
    echo "   restart          - 全てのコンテナを再起動"
    echo "   build            - キャッシュを無視して全コンテナのイメージをビルド"
    echo "   ps               - コンテナの稼働状況を確認"
    echo ""
    echo " [ログ監視]"
    echo "   logs-middleware  - Middleware (仮想ミドルウェア) のログを監視"
    echo "   logs-processing  - Processing Unit (処理ユニット) のログを監視"
    echo ""
    echo " [永続化・Valkey操作]"
    echo "   valkey-keys      - ホスト上のValkeyに保存されている全キーを表示"
    echo "   valkey-cli       - ホスト上のValkey-cliに接続"
    echo ""
    echo " [テスト・ビルド]"
    echo "   test             - Goプロジェクト全体の単体テストを実行"
    echo "   proto            - protoファイルからGoのgRPCコードを生成"
    echo "======================================================================"
}

COMMAND=$1

case "$COMMAND" in
    up)
        docker-compose up -d
        ;;
    down)
        docker-compose down
        ;;
    restart)
        docker-compose restart
        ;;
    build)
        docker-compose build --no-cache
        ;;
    ps)
        docker-compose ps
        ;;
    logs-middleware)
        docker-compose logs -f middleware
        ;;
    logs-processing)
        docker-compose logs -f processing
        ;;
    valkey-keys)
        if command -v valkey-cli &> /dev/null; then
            valkey-cli KEYS "*"
        elif command -v redis-cli &> /dev/null; then
            redis-cli KEYS "*"
        else
            echo "Error: valkey-cli or redis-cli not found."
            exit 1
        fi
        ;;
    valkey-cli)
        if command -v valkey-cli &> /dev/null; then
            valkey-cli
        elif command -v redis-cli &> /dev/null; then
            redis-cli
        else
            echo "Error: valkey-cli or redis-cli not found."
            exit 1
        fi
        ;;
    test)
        go test -v ./...
        ;;
    proto)
        protoc --go_out=. --go_opt=paths=source_relative \
            --go-grpc_out=. --go-grpc_opt=paths=source_relative \
            proto/processing.proto
        ;;
    *)
        show_help
        ;;
esac
