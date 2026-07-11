# ローカル環境規約

## introduction
- ここにはローカル環境に関して記述しています


## 構成
- **アーキテクチャ・設計思想**に記載されているとおりスペースベースアーキテクチャにもとづいたシステム構成となっています。
　- **Processing Unit (処理ユニット)**:、**Processing Unit (処理ユニット)**、**Valkey (永続化・キャッシュ)**が存在する。
  - **Processing Unit (処理ユニット)**:、**Processing Unit (処理ユニット)**については**Docker**で動かす
  - **Valkey (永続化・キャッシュ)**についてはローカル環境で既に起動している前提とする


## ローカル環境構築
- ユーザからローカル環境の起動指示があった場合、以下の手順を行う（コマンドを実行する）
    - 実行手順1〜4に記載されたコマンドは実行しないこと。
- 実行手順
    - 1. 既存のDocker停止：`sudo docker compose down`
    - 2. イメージビルド：`sudo docker compose build`
        - Dockerファイルにテスト実行を定義しています。テストでNGとなった場合は詳細を出力して中断する
    - 3. イメージデプロイ：`sudo docker compose up -d`
    - 4. イメージ実行確認：`sudo docker ps`
        - **pacebase-processing:latest**と**pacebase-processing:latest**のIMAGEが起動していることを確認する
- 1〜4の実行中で予期せぬエラーが発生した場合は中断してエラーの内容を伝えること
