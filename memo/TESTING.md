# ローカルでの動作確認方法

このドキュメントでは、ローカル環境で Docker を使用してアプリケーションの動作確認を行う方法をご説明します。

## 1. Docker を使用したローカル実行
- 以下のコマンドでコンテナイメージをビルドします:

  docker build -t spacebase .

- 以下のコマンドでコンテナを起動します:

  docker run -p 8080:8080 spacebase

## 2. API エンドポイントのテスト例
- **プロファイルの作成**:
  
  curl -X POST http://localhost:8080/profiles \\
    -H "Content-Type: application/json" \\
    -d '{"id":101,"name":"Alice","status":"active","languages":["Go","Python"]}'

- **プロファイルの取得**:
  
  curl http://localhost:8080/profiles/101