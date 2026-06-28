# 仮想ミドルウェア構築計画

このドキュメントでは、Space-Based Architecture の「仮想ミドルウェア層」を具体的に実装するための計画を示します。

## 1. 要件整理
- 処理ユニット同士の疎結合連携を実現  
- タプルスペース（Tuple Space）または Key–Value Grid 操作を抽象化  
- スケールアウト／クラスタ環境でも一貫したアクセスを提供  
- **ミドルウェア層で API ゲートウェイ機能を提供し、外部パブリックAPIを公開**

## 2. インターフェース設計
- `Space` インターフェース定義  
  ```go
  type Tuple struct {
    Key   string
    Value []byte
    TTL   time.Duration
  }
  type Space interface {
    Put(ctx context.Context, tuple Tuple) error
    Get(ctx context.Context, key string) (Tuple, error)
    Keys(ctx context.Context, pattern string) ([]string, error)
  }
  ```
- エラー処理・再試行ポリシーをインターフェースに含める検討

## 2.1 API Gateway インターフェース設計
- ミドルウェア層が公開するパブリックAPIを定義  
  ```go
  type Gateway interface {
    ServeHTTP(http.ResponseWriter, *http.Request)
    Route(ctx context.Context, path string, params map[string]string) (Response, error)
  }
  ```
- 認証・認可、レートリミット、ロギングなど共通機能を組み込み  
- ルーティングロジックはユーザIDやパスパラメータに基づき、対象 PU へ転送

## 2.2 同期・非同期モードサポート
- Gateway に設定可能なモード（Sync／Async）を追加  
  - **Sync モード**：HTTP/gRPC で即時フォワード  
  - **Async モード**：メッセージブローカー（Kafka/RabbitMQ/PubSub）へイベントを投げる  
- Strategy パターンで `Gateway` 実装を分離し、設定ファイルまたは環境変数で切り替え可能  
- Config 例（YAML/環境変数）:  
  ```yaml
  gateway:
    mode: "sync"  # or "async"
  ```

## 3. アダプタ層実装
- Redis（ValKey）のドライバ向けアダプタ  
  - `RedisSpace` 構造体を実装  
- 将来的に Hazelcast／Ignite／GigaSpaces 用アダプタを追加可能に構造化

## 4. 実装フェーズ
1. `internal/space/space.go` に Space インターフェース定義  
2. `internal/space/redis_adapter.go` に Redis アダプタを実装  
3. `internal/gateway/gateway.go` に Gateway／SyncGateway／AsyncGateway を実装  
4. `internal/handler`／`internal/service` から直接 Redis 呼び出し箇所を Space/Gateway 経由に切り替え  
5. DI（依存性注入）またはファクトリーパターンでモード別 Gateway を選択可能に

## 5. テスト戦略
- ユニットテストでモック `Space`／`Gateway`（Sync/Async）を用意  
- Redis 実環境への統合テスト  
- 非同期イベントトリガーのエンドツーエンドテスト  
- 負荷テスト（複数 PU から同時 Put/Get、API 呼び出し分散）

## 6. デプロイ
- Dockerfile にミドルウェア実装を含めて再ビルド  
- ミドルウェア（Space 層 + Gateway）を独立コンテナまたは PU と同一コンテナに組み込み  
- 複数ノードでクラスタ構成を検証し、同期／非同期モードを切り替えて挙動確認

## 7. 次工程
- イベント駆動連携（Pub/Sub）機能強化  
- キャッシュポリシー／Write-Behind などの高度機能検討  
- ユーザIDベースのインスタンス割り振り（Hash(userID)%N）ルーティング実装

以上のステップで、同期・非同期両モードを切り替え可能な API ゲートウェイ付き仮想ミドルウェア層を実装します。