# ローカル環境規約

## 構成
- **Middleware**、**Processing Unit**：Dockerコンテナで動作。
- **Valkey**：ローカル環境で起動している前提とする。

## ローカル環境操作（Makefile優先）
ローカル環境の操作は、プロジェクトルートにある `Makefile` のコマンドを最優先で使用すること。
- **ビルド・起動**: `make up` または `make build`
- **停止**: `make down`
- **ログ確認**: `make logs-middleware` または `make logs-processing`
- **テスト実行**: `make test`

## docker compose を直接使用する場合の起動手順
ユーザーから明示的な指示がある場合のみ、以下の手順を実行する。
1. **停止**: `docker compose down`
2. **ビルド**: `docker compose build` （テスト失敗時は中断）
3. **起動**: `docker compose up -d`
4. **起動確認**: `docker ps` で `spacebase-middleware:latest` と `spacebase-processing:latest` が起動しているか確認。
