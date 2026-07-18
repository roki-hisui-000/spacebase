# ローカル環境規約

## 構成
- **Processing Unit**: Dockerコンテナで起動。
- **Valkey**: ローカル環境（ホスト上）で起動済みとする。

## ローカル環境起動手順
- ユーザーから明示的な起動指示があった場合のみ実行する（勝手に起動しない）。
- 起動手順：
  1. コンテナ停止: `make down` (または `docker compose down`)
  2. ビルド: `make build` (または `docker compose build`)
     - テストNG時はエラー詳細を出力して処理を中断。
  3. 起動: `make up` (または `docker compose up -d`)
  4. 確認: `docker ps` で `spacebase-processing` コンテナの起動を確認。
- 予期せぬエラー発生時は処理を中断し、エラー内容を報告すること。
