# ローカル環境規約

## 構成
- **Processing Unit**: Dockerコンテナで起動。
- **Middleware**: Dockerコンテナで起動。
- **Valkey**: Dockerコンテナで起動。Docker Compose のライフサイクル内で自動管理。

## ローカル環境起動手順
- ユーザーから明示的な起動指示があった場合のみ実行する（勝手に起動しない）。
- **起動前の前提条件（初回クローン時）：**
  1. **ツールセットアップ:** `make setup` を実行して、必要な開発ツール（Go, protocなど）をインストールします（miseが未インストールの場合、インストールスクリプトが案内されます）。
     - ※自動生成コード（`*.pb.go`）は、`make build` や `make up` 実行時に内部で `make proto` が自動的に走り、毎回最新の状態で生成されるため、個別にコマンドを実行する必要はありません。
- 起動手順：
  1. コンテナ停止: `make down` (または `docker compose down`)
  2. ビルド: `make build` (または `docker compose build`)
     - テストNG時はエラー詳細を出力して処理を中断。
  3. 起動: `make up` (または `docker compose up -d`)
  4. 確認: `docker ps` で `spacebase-processing` コンテナの起動を確認。
- 予期せぬエラー発生時は処理を中断し、エラー内容を報告すること。
