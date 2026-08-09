# [ダッシュボード作成] バックエンド設計・API仕様

[issue](https://github.com/roki-hisui-000/spacebase/issues/31)

## 概要
管理者ダッシュボードに関するバックエンド処理（永続化およびデータ取得）の設計・API仕様です。

## バックエンドAPI仕様

### 1. 注文情報の永続化 API
外部からの注文リクエスト情報を Valkey/Redis へ永続化します。

- **URL**: `/api/dashboard/orders`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "userId": "string",
    "price": 100,
    "status": "string", // completed | order | reject | error
    "requestId": "string"
  }
  ```
- **処理ロジック**:
  1. `orderId` は `ord_<UUID>` の形式で自動採番
  2. `createdAt` には現在時刻（UTC）を自動設定
  3. Valkey/Redisにキー `"order:<orderId>"` で永続化
- **Response Body**: 保存された `RecentOrder` オブジェクト
  ```json
  {
    "orderId": "ord_89a3f...",
    "userId": "user_42",
    "price": 150,
    "status": "completed",
    "requestId": "req_1",
    "createdAt": "2026-07-04T09:55:02Z"
  }
  ```

### 2. ダッシュボードデータ取得 API
ダッシュボード全体の情報を返します。リアルタイム配信をサポートしています。

- **URL**: `/api/dashboard`
- **Method**: `GET`
- **Query Parameter**:
  - `stream` (boolean, optional): `true` を指定、または `Accept: text/event-stream` ヘッダーがある場合、SSEによるリアルタイムデータストリーミング配信を行います（3秒間隔）。
- **Response Body (通常時)**:
  ```json
  {
    "workerStatus": "running",
    "metrics": {
      "currentStock": 0,
      "queueLength": 65,
      "dbOrderCount": 35
    },
    "recentOrders": [
      { "orderId": "ord_89a3f...", "userId": "user_42", "price": 150, "status": "completed", "requestId": "req_1", "createdAt": "2026-07-04T09:55:02Z" }
    ]
  }
  ```
- **Response Body (SSEストリーム時)**:
  `Content-Type: text/event-stream` でチャンク送信。
  ```text
  data: {"workerStatus":"running","metrics":{...},"recentOrders":[...]}
  
  ```
