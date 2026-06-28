# 仮想ミドルウェア層アーキテクチャ設計

## 1. 概要
Space-Based Architecture の「仮想ミドルウェア層」は、以下の機能を提供します。
- タプルスペース／Key–Value Grid 操作の抽象化  
- API ゲートウェイ機能（認証・認可／レートリミット／ログ）  
- 同期モード（Sync）／非同期モード（Async）の切り替え可能  

## 2. コンポーネント図
```text
┌───────────────┐      ┌───────────────┐
│  Client/API   │◀────▶│    Gateway    │
│ (HTTP, gRPC)  │      │ (Sync/Async)  │
└───────────────┘      └──────┬────────┘
                              │Route
                              ▼
                       ┌───────────────┐
                       │    Space      │
                       │(Put/Get/Keys) │
                       └──────┬────────┘
                              │Adapter
                              ▼
                       ┌───────────────┐
                       │    Redis      │
                       │  (Key–Value)  │
                       └───────────────┘
```

## 3. パッケージ構成
```
internal/
├─ space/
│  ├─ space.go          // Space インターフェース定義
│  └─ redis_adapter.go  // RedisSpace 実装
└─ gateway/
   ├─ gateway.go        // Gateway インターフェース & HTTP ハンドラ
   ├─ sync_gateway.go   // SyncGateway 実装
   └─ async_gateway.go  // AsyncGateway 実装（Kafka/RabbitMQ など）
```

## 4. データフロー
1. クライアントが Gateway の HTTP/gRPC エンドポイントを呼び出し  
2. Gateway が認証・認可・レート制御・ロギングを実施  
3. Route メソッドでパスとパラメータをもとに処理ユニット（PU）を選定  
4. Sync モード: Space.Put/Get を直接呼び出し  
5. Async モード: Space.Put 相当のイベントをメッセージブローカーに発行  

## 5. コンフィグレーション例
```yaml
gateway:
  mode: "sync"   # or "async"
space:
  redis:
    address: "localhost:6379"
    password: ""
    db: 0
```

## 6. 今後の拡張
- Hazelcast／Ignite アダプタの追加  
- Write-Behind キャッシュポリシー  
- ユーザIDベースのインスタンス割り振りルーティング