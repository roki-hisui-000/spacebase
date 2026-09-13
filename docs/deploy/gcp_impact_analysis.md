# GCPデプロイにおける影響箇所と構成変更の洗い出し

本ドキュメントでは、現在のSpacebaseシステム構成（Go + gRPC + HTTP + Valkey）をGoogle Cloud Platform (GCP) へ移行・デプロイする際の影響箇所、考慮事項、および推奨アーキテクチャについてまとめています。

---

## 1. システム概要と現状の構成
現在のローカル環境（Docker Compose）のサービス構成は以下の通りです。
- **Middleware**: 外部HTTP受付 (`:8080`) と、Processing Unit向けおよび外部向けのgRPCサーバー (`SpaceService`, `:50052`)。唯一Valkeyに直接接続する。
- **Processing Unit**: ビジネスロジックを処理するgRPCサーバー (`:50051`)。`Middleware` の `SpaceService` を経由してデータを操作。
- **Valkey**: Redis 8互換のインメモリデータストア (`:6379`)。コンテナ再起動時も `./data/valkey` へのAOF+RDBマウントによりデータを永続化。

---

## 2. カテゴリ別の影響箇所と課題

### 2.1 サービスアーキテクチャ・通信方式の影響

#### ① Middlewareのマルチポート・マルチプロトコル公開
- **現状**: `Middleware` は、クライアント向けのHTTP/RESTAPI (`:8080`) と、内部通信用のgRPC (`:50052`) の両方を同一のコンテナインスタンスで同時に起動しています。
- **GCPでの影響**:
  - **Cloud Runの場合**: Cloud Runは原則として1つのサービスコンテナに対して単一のポート（通常 `PORT` 環境変数で指定されるポート、デフォルト8080）へのリクエスト転送を前提としています。HTTPとgRPCの複数ポートを同時にパブリックに公開、または個別のエンドポイントで受けるのは難しいため、HTTPゲートウェイとgRPCサービスを別々のコンテナ（または別サービス）に分割するリファクタリング、あるいはプロキシ（Envoy等）の導入が必要になります。
  - **GKE (Google Kubernetes Engine) / GCE (Compute Engine) の場合**: ServiceリソースやVM側のポートマッピングを利用して複数ポート（`:8080` と `:50052`）を同時に問題なく公開できます。コード変更は不要です。

#### ② 双方向gRPC通信と名前解決（サービスディスカバリ）
- **現状**: `Middleware` と `Processing Unit` は、Docker Composeのブリッジネットワークを介して互いに `processing:50051` や `middleware:50052` といったコンテナ名で相互に通信しています。
- **GCPでの影響**:
  - **GKEの場合**: CoreDNSが標準搭載されているため、K8s内のService名（例: `http://processing-service.default.svc.cluster.local:50051`）で極めて容易に名前解決が可能です。
  - **Cloud Runの場合**: Cloud Runはエンドポイントがそれぞれ独立した `https://*.run.app` になります。Cloud Run同士の相互呼び出しでは、それぞれの外部URL、あるいは「サーバーレスVPCアクセス」等を経由してプライベート通信を行うように通信先環境変数（`SPACE_SERVICE_ADDR`, `PROCESSING_UNIT_ADDR`）を調整する必要があります。
  - **非暗号化gRPC**: 現在 `grpc.Dial(..., grpc.WithInsecure())` で通信を平文化していますが、Cloud Runや外部公開ロードバランサを挟む場合は、HTTPS (h2c) による通信暗号化を前提とした実装、または認証トークン（OIDC IDトークン）の付与が必要になる可能性があります。

---

### 2.2 データストア（Valkey）の影響と選択肢

#### ① 永続化ストレージとインフラの選定
- **現状**: ホストの物理ディスク `./data/valkey` にマウントしてAOF+RDBによるデータ永続化を行っています。
- **GCPでの影響**:
  - **選択肢A: Memorystore for Redis (推奨)**
    - GCPのフルマネージドRedisサービスを利用。耐障害性、自動バックアップ、高可用性（M/S構成）がクラウド側で保証されます。
    - *影響*: MemorystoreはVPC内のプライベートIPのみ提供するため、アプリケーション側（Cloud RunならVPCアクセス用のコネクタ設定、GKEなら同一VPC内でのデプロイ）でネットワーク設計が必要になります。また、環境変数 `REDIS_HOST` にMemorystoreのプライベートIP、`REDIS_PORT` に接続ポート（通常6379）を注入する必要があります。
  - **選択肢B: GKEでのStatefulSetによるセルフホストValkey**
    - Kubernetes上にValkeyコンテナをデプロイし、PersistentVolume (GCPのPersistent Diskなど) をマウントします。
    - *影響*: 永続化を維持するため、`docs/valkey_persistence.md` に提示されているような `StatefulSet` + `PersistentVolumeClaim` (PVC) のマニフェスト整備が必要になります。

---

### 2.3 設定管理・セキュリティの影響

#### ① Secret Managerの連携
- **現状**: `config/config.go` にすでに `os.Getenv("SECRET_NAME")` がある場合に Google Cloud Secret Manager からJSON設定を動的に読み込むロジックが組み込まれています。
- **GCPでの影響**:
  - デプロイ先の実行環境（GKE Pod、Cloud Run、GCEインスタンス）が、Secret Managerから値を読み取れる適切なGCPサービスアカウント（GSA）と、それに紐づくIAM権限（`roles/secretmanager.secretAccessor`）を付与されている必要があります。
  - Secret Managerに格納するJSONフォーマットを `config/config.go` の `Config` 構造体 (`redis_host`, `redis_port` 等) と正確に一致させる必要があります。

#### ② 各種環境変数の注入
- ローカル環境で使用している環境変数を、デプロイ基盤側の環境変数としてマウント・注入する仕組みを整備します。
  - `SPACE_SERVICE_ADDR`
  - `PROCESSING_UNIT_ADDR`
  - `REDIS_HOST`, `REDIS_PORT` (Secret Managerを使用しない場合のフォールバック用)

---

### 2.4 コンテナビルドとCI/CDの影響

#### ① コンテナレジストリとビルドアーキテクチャ
- **現状**: 各ディレクトリ内の `Dockerfile`（`golang:1.25` → `debian:stable-slim`）から手動、または `Makefile` (`make build`) でビルドしています。
- **GCPでの影響**:
  - ビルドしたイメージをGCPのコンテナレジストリである **Artifact Registry** にプッシュするパイプライン（GitHub Actions、あるいは GCP Cloud Build）を構築する必要があります。
  - ビルド時のプロセッサアーキテクチャ（amd64 / arm64）を実行基盤のCPUに合わせる必要があります。

#### ② ヘルスチェック
- コンテナの稼働状況を監視するため、GCPロードバランサ、Cloud RunのStartup/Livenessプローブ、あるいはGKEプローブが、Middleware（HTTP `:8080`、gRPC `:50052`）やProcessing Unit（gRPC `:50051`）に対してヘルスチェックを実行できるようにする必要があります。
  - HTTP側に `/healthz` や `/ping` などの疎通確認用エンドポイントの追加。
  - gRPC標準のヘルスチェックプロトコル（gRPC Health Checking Protocol）の導入検討。

---

## 3. 推奨されるGCPホスティング構成の比較

| 構成案 | メリット | デメリット / アプリケーション側への影響 |
| :--- | :--- | :--- |
| **A. GKE (Google Kubernetes Engine)** <br>*(最も推奨・相性が良い)* | - 同一Pod/Service内でHTTP/gRPCの複数ポート公開が極めて容易。<br>- CoreDNSによるサービス間名前解決が標準機能として動作。<br>- ValkeyのStatefulSetセルフホストも選択可能。 | - K8sマニフェスト（Deployment, Service, StatefulSet等）の作成と管理コスト。<br>- クラスタの維持費および運用負荷。 |
| **B. Cloud Run** <br>*(サーバーレス・低運用コスト)* | - インフラ運用の手間がほぼゼロ。<br>- 使用した分だけの従量課金、オートスケーリング。<br>- Secret Manager等との統合が非常に容易。 | - Middlewareのマルチポート・プロトコル同時起動が困難なため、HTTPゲートウェイとgRPCサービスを別コンテナ（別サービス）に分離するコードの変更・分割が必要。<br>- コールドスタートや、VPCコネクタ経由のValkey（Memorystore）接続設定が必要。 |
| **C. Compute Engine (GCE) + Docker Compose** <br>*(最速・最小変更)* | - 現在の `docker-compose.yml` をほぼそのまま移行可能。<br>- 最もアプリケーション側の修正が少ない。 | - 自動スケーリングやゼロスケールなどのクラウドマネージドの恩恵が得られない。<br>- VM自体の監視・OSパッチ当てなどの運用管理負荷が残る。 |

---

## 4. 今後の移行ロードマップ推奨アクション

1. **ホスティング基盤の決定**: 
   - 運用の簡便性とコストパフォーマンス重視なら「Cloud Run（Middlewareのポート分割、またはEnvoyでのルーティングプロキシを導入）」、現状のコード変更を最小化し、将来の拡張性を担保するなら「GKE」を選定。
2. **ネットワーク・Valkeyの設計**: 
   - 永続化を考慮し「Memorystore for Redis」を構成し、VPC/ネットワークアクセスを確保する。
3. **Secret Manager / 環境変数の整理**:
   - GCP上にSecretを作成し、各アプリケーションの接続情報を安全に供給できるようにする。
4. **CI/CD・Artifact Registryの構築**:
   - `Dockerfile` のビルド・テスト・プッシュを自動化するパイプラインを構築する。

---

## 結論
Cloud Run vs GKEについて記載されていますが、今回はCloud Runを利用します。
