# API一覧 - Spacebase

このドキュメントは、Spacebase プロジェクトで利用可能な API の一覧を示します。

## エンドポイント一覧

### 1. ユーザ登録 API

- **エンドポイント:** `/register`
- **HTTPメソッド:** POST  
- **説明:** 新しいユーザプロファイルを永続化し、「User registered successfully」のメッセージを返します。  
- **リクエストボディ:** JSON形式。例:
  ```json
  {
    "name": "Roki",
    "status": "active",
    "languages": ["Go","Java"]
  }
  ```
- **レスポンス:** HTTPステータスコード 200 OK  
  レスポンスボディにはメッセージ文字列を返却します。
  ```text
  User registered successfully
  ```

### 2. ユーザプロファイル取得 API（未実装）

- **エンドポイント:** `/profiles/{id}`
- **HTTPメソッド:** GET  
- **説明:** 指定したIDのユーザプロファイルを取得します。現時点では未実装のため 404 を返します。  
- **URLパラメータ:**  
  - `id` – ユーザプロファイルの識別子  

## アクセス方法

Docker コンテナはホストネットワークモードで稼働し、ホストのポート 8080 でリッスンしています。  
以下の URL を使用して API にアクセスしてください:

- ユーザ登録 API:  
  ```
  http://localhost:8080/register
  ```
- ユーザプロファイル取得 API:  
  ```
  http://localhost:8080/profiles/{id}
  ```

### curl 例

ユーザ登録:
```
curl -i -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"id":100,"name":"Roki","status":"active","languages":["Go","Java"]}'