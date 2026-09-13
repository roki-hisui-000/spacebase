# GCPデプロイメモ

## 前提
- cloud runを利用して実現する

## 現在構成（ローカル向け）

- Middleware
	- 外部に唯一の公開
	- Process Unitを呼び出して処理を委譲
	- データベースにアクセス
- Processing Unit
	- ビジネスロジックを担当
	- データベースにアクセスする際はMiddlewareを呼び出す
- データベース
	- valkeyを使用

### ポート番号について

- `8080`：外部→Middleware
- `50051`：Middleware→Processing Unit（処理の委譲）
- `50052`：Processing Unit→Middleware（主にデータベースへのアクセス呼び出し）
- `6379`：Middleware→データベース

## デプロイにおける課題

### 通信経路

1. **Middleware**はローカル環境で`8080`と`50052`を利用している。Cloud runでは単一ポートしか公開できない。
2. **Middleware**と**Processing Unit**はDocker composeのブリッジネットワークを利用して`<コンテナ名>:<ポート>`で通信しているが、<br>
	Cloud Runではエンドポイントを独立したURLで通信する必要がある。

### データベース

1. データベース（Valkey）は`/data/valkey`配下に永続化したデータを保存している。これは、データベースが停止しても永続化したデータが消去されない措置です。<br>
   GCPで同様のふるまいをするには、GCPフルマネージドRedisサービスを利用する必要がある。

### セキュリティ

1. アプリ起動時に、GCPのSecret ManagerからJSON設定を読み込む実装がされているので、Secret Managerに秘匿情報を設定する必要がある。
2. 環境変数を読み込む設定を導入する
	- 現在の実装では、ローカルでは`docker-compose.yml`に定義しているので以下に変更する
		- 上位環境は**Secret Manager**
		- 下位環境では **.env**に定義する。

## CI/CDの影響

1. ローカルでビルドしたDockerのイメージをGCPのコンテナレジスト（Artifact Registry）にプッシュするパイプライン（GCP Cloud Build）を構築する必要がある
	- イメージ作成時には実行基盤のCPU（例：arm64 / amd64）に合わせる必要がある
2. コンテナの稼働状況を監視するためヘルスチェックを導入する



## 課題の対応方針

1. **Middleware**を以下の2つのサービスに分ける
	- HTTPゲートウェイ：外部からのリクエストを受け付ける
	- gRPCゲートウェイ：主にデータベースにアクセスするためにProcessing Unitから呼ばれる

2. **Middleware**と**Processing Unit**間の通信を確立するため通信先環境変数を用意する
	- 通信先環境変数はcloud runのコンテナ環境変数（`--set-env-vars`で設定）
	- 設定例：
		- Middleware 側の環境変数: `PROCESSING_UNIT_ADDR=processing-unit-xxxx-an.a.run.app:443`
		- Processing Unit 側の環境変数: `SPACE_SERVICE_ADDR=middleware-space-xxxx-an.a.run.app:443`
	- 以下に記載したとおり、実装とインフラを変更する
		1. **コード側**: `grpc.WithInsecure()` から、**TLS暗号化 (`credentials.NewTLS`) への対応**
		2. **コード側**: Google Cloud の **IDトークン（IAM認証ヘッダー）自動付与ロジック** の追加
		3. **インフラ側**: gRPCサービスの Cloud Run デプロイ時に **`--use-http2` を指定**
		4. **インフラ側**: 呼び出し元サービスアカウントに **`roles/run.invoker` 権限を付与**

3. データベース
	- Cloud Run から VPC の中へ安全に入り込むためのトンネル（Direct VPC 下り または VPCアクセス コネクタ）を開通させる必要がある
	- 手順
		1. Memorystore for Redis インスタンスの作成
			- GCPコンソールまたは `gcloud` コマンドで Redis インスタンスを作成する
		2. Cloud Run と VPC を繋ぐネットワーク設定（Direct VPC 下り）
			```
			# デフォルトVPC (default) の東京リージョンに 1GB の Redis を作成
			gcloud redis instances create spacebase-redis \
		    --size=1 \
		    --region=asia-northeast1 \
		    --zone=asia-northeast1-a \
		    --redis-version=redis_7_0 \
		    --network=projects/YOUR_PROJECT_ID/global/networks/default
		    ```
		3. 接続先情報を Secret Manager（または環境変数）に登録
			```
			{
				"redis_host": "10.0.0.4",
				"redis_port": "6379",
				"app_port": "8080",
				"redis_ttl": "24h",
				"admin_user": "admin",
				"admin_pass": "強固なパスワード"
			}
			```
		4. Cloud Run（Space Data サービス）のデプロイ
			```
			gcloud run deploy spacebase-space-service \
			    --image asia-northeast1-docker.pkg.dev/YOUR_PROJECT/repo/spacebase:latest \
			    --region asia-northeast1 \
			    --no-allow-unauthenticated \
			    # --- ネットワーク設定 (VPC直結) ---
			    --network default \
			    --subnet default \
			    --vpc-egress private-ranges-only \
			    # --- 環境変数設定 ---
			    --set-env-vars "SECRET_NAME=projects/YOUR_PROJECT/secrets/spacebase-app-config/versions/latest"
			```
4. セキュリティ
	- Secret Managerから値を取得できるように`roles/secretmanager.secretAccessor`権限を付与する

5. CI/CD

	- コンテナの稼働状況を監視するヘルスチェックを導入する
		1. Cloud Run デプロイ時（または YAML マニフェスト）で、HTTPプローブを有効化します。
		```bash
			gcloud run deploy spacebase-gateway \
			--image asia-northeast1-docker.pkg.dev/YOUR_PROJECT/repo/spacebase-gateway:latest \
			--port 8080 \
			# --- Startupプローブ（起動完了検知）---
			--startup-probe httpGet.path=/health,httpGet.port=8080,initialDelaySeconds=2,periodSeconds=5,failureThreshold=3 \
			# --- Livenessプローブ（死活監視）---
			--liveness-probe httpGet.path=/health,httpGet.port=8080,periodSeconds=10,failureThreshold=3
		```
		2. gRPCサービスの手順
			- ヘルスチェック用のgRPCライブラリ（`google.golang.org/grpc/health`）を導入する
				- 以下の実装を追加してヘルスチェックを有効にする
					```
					// ★ ここを追加：標準ヘルスチェックサービスを登録
					healthServer := health.NewServer()
					healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
					healthpb.RegisterHealthServer(grpcServer, healthServer)
				    ```
			- Cloud Run 側で gRPC ヘルスチェックを指定してデプロイ
				```bash
					gcloud run deploy spacebase-processing \
					--image asia-northeast1-docker.pkg.dev/YOUR_PROJECT/repo/spacebase-processing:latest \
					--port 50051 \
					--use-http2 \
					# --- gRPC Startupプローブ ---
					--startup-probe grpc.port=50051,initialDelaySeconds=2,periodSeconds=5,failureThreshold=3 \
					# --- gRPC Livenessプローブ ---
					--liveness-probe grpc.port=50051,periodSeconds=10,failureThreshold=3
				```

## 補足

- Dockerイメージはローカル/GCPで共通したものを利用する
	- 環境別の設定ファイルはローカルの場合は".env"、GCPではSecret Managetの値を参照する

