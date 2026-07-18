.PHONY: help up down restart build ps logs-middleware logs-processing valkey-keys valkey-cli test proto setup

# デフォルトターゲット：ヘルプ表示
help:
	@echo "======================================================================"
	@echo " Spacebase Development Skills Commands (Makefile)"
	@echo "======================================================================"
	@echo " 開発やAIエージェントの操作をサポートするための便利なコマンド集です。"
	@echo ""
	@echo " [環境構築・ツール管理]"
	@echo "   make setup            - miseツールマネージャーと必要な開発ツール一括セットアップ"
	@echo ""
	@echo " [コンテナ操作]"
	@echo "   make up               - 全てのコンテナをバックグラウンドで起動"
	@echo "   make down             - 全てのコンテナを停止・削除"
	@echo "   make restart          - 全てのコンテナを再起動"
	@echo "   make build            - キャッシュを無視して全コンテナのイメージをビルド"
	@echo "   make ps               - コンテナの稼働状況を確認"
	@echo ""
	@echo " [ログ監視]"
	@echo "   make logs-middleware  - Middleware (仮想ミドルウェア) のログを監視"
	@echo "   make logs-processing  - Processing Unit (処理ユニット) のログを監視"
	@echo ""
	@echo " [永続化・Valkey操作]"
	@echo "   make valkey-keys      - ホスト上のValkeyに保存されている全キーを表示"
	@echo "   make valkey-cli       - ホスト上のValkey-cliに接続"
	@echo ""
	@echo " [テスト・ビルド]"
	@echo "   make test             - Goプロジェクト全体の単体テストを実行"
	@echo "   make proto            - protoファイルからGoのgRPCコードを生成"
	@echo "======================================================================"

# 全てのコンテナを起動
up:
	docker-compose up -d

# 全てのコンテナを停止・削除
down:
	docker-compose down

# 全てのコンテナを再起動
restart:
	docker-compose restart

# コンテナイメージのビルド
build:
	docker-compose build --no-cache

# コンテナの稼働状況確認
ps:
	docker-compose ps

# 仮想ミドルウェアのログ監視
logs-middleware:
	docker-compose logs -f middleware

# 処理ユニットのログ監視
logs-processing:
	docker-compose logs -f processing

# Valkeyに保存されている全キーの確認
valkey-keys:
	valkey-cli KEYS "*" || redis-cli KEYS "*"

# Valkey CLIの起動
valkey-cli:
	valkey-cli || redis-cli

# 単体テストの実行
test:
	go test -v ./...

# gRPCプロトコルのコード生成
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/processing.proto

# miseのセットアップと定義ツールの自動インストール
setup:
	@if ! command -v mise >/dev/null 2>&1; then \
		echo "======================================================================"; \
		echo " mise がインストールされていません。インストールを開始します..."; \
		echo "======================================================================"; \
		curl https://mise.run | sh; \
		echo ""; \
		echo "👉 mise のインストールが完了しました！"; \
		echo "お使いのシェルをアクティベートするため、以下を ~/.bashrc もしくは ~/.zshrc 等に追記してください："; \
		echo "  echo 'eval \"\$$($$HOME/.local/share/mise/bin/mise activate bash)\"' >> ~/.bashrc"; \
		echo "  (※zshの場合は bash を zsh に変更してください)"; \
		echo "======================================================================"; \
	else \
		echo "mise は既にインストールされています。"; \
	fi
	@echo "定義されたツール群（Go, protoc, プラグイン）をインストールします..."
	@if [ -f "$$HOME/.local/bin/mise" ]; then \
		$$HOME/.local/bin/mise install; \
	elif [ -f "$$HOME/.local/share/mise/bin/mise" ]; then \
		$$HOME/.local/share/mise/bin/mise install; \
	elif command -v mise >/dev/null 2>&1; then \
		mise install; \
	else \
		echo "mise のパスが見つかりません。シェルを再起動するか、手動で path を通してください。"; \
	fi


# ==============================================================================
# レビュー用タスク (トークン削減対策)
# ==============================================================================

# 1. developブランチ（追跡ブランチ）との差分を、不要なファイルを除外して出力する
diff-review:
	git diff origin/develop..HEAD -- . ':!*.pb.go' ':!*.html' ':!*.yaml' ':!*.json'

# 2. まだコミットしていないローカルの変更（作業中）を、不要なファイルを除外して出力する
diff-review-local:
	git diff HEAD -- . ':!*.pb.go' ':!*.html' ':!*.yaml' ':!*.json'
