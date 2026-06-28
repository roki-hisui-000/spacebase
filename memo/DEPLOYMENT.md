# GCP Cloud Run デプロイ手順


## 事前準備
- 設定ファイルを準備  
  - config/config.dev.yaml、config/config.stg.yaml、config/config.prod.yaml に `<TODO>` プレースホルダーを設定  
  - 環境変数 `GO_ENV`（dev、stg または prod）を指定し、該当する設定ファイルをロード  

## ローカルでの起動
ローカル環境では Docker コンテナを直接起動し、環境変数で動作を制御できます。例として以下のコマンドを実行してください。

```bash
sudo docker run -d --name spacebase-app \
  -p 8080:8080 \
  -e REDIS_HOST=127.0.0.1 \
  -e REDIS_PORT=6379 \
  -e APP_PORT=8080 \
  -e REDIS_TTL=24h \
  -e VALKEY="<実際のValKey>" \
  spacebase-server:latest
```

これにより、config パッケージは環境変数を読み込み、同じ設定項目を適用してアプリケーションを起動します。

## 必須環境変数
| 変数名 | 説明                                    | デフォルト値 |
| ------ | --------------------------------------- | ------------ |
| GO_ENV | ロードする設定ファイルの環境 (dev/stg/prod) | dev          |

※Redis や ValKey の情報は設定ファイル内で管理します。

### 設定項目説明
- **GO_ENV**: `dev`、`stg` または `prod` を指定し、対応する YAML ファイルをロードします。
- **REDIS_HOST/REDIS_PORT**: Redis 接続先ホスト名とポート。YAML → 環境変数 → コード内デフォルトの順でフォールバック。  
- **APP_PORT**: サーバー起動ポート。YAML → 環境変数 → デフォルト `8080`。  
- **REDIS_TTL**: Redis データ有効期限。YAML → 環境変数 → デフォルト `24h`。  
- **VALKEY**: アプリケーションが Key-Value ストア（ValKey）へのアクセスやデータ暗号化・署名に用いる認証キー／シークレットです。ローカルではダミー値を、運用環境では Secret Manager 等で管理してください。

---

## 1. Artifact Registry のセットアップ
1. GCP プロジェクト ID: `<TODO_PROJECT_ID>`  
2. リージョン: `<TODO_REGION>`  
3. リポジトリ: `<TODO_REPOSITORY>`  
4. Artifact Registry API を有効化  
5. Docker 認証  
```bash
gcloud auth configure-docker <TODO_REGION>-docker.pkg.dev
```

## 2. Docker イメージのビルド
```bash
docker build -t <TODO_REGION>-docker.pkg.dev/<TODO_PROJECT_ID>/<TODO_REPOSITORY>/spacebase:latest .
```

## 3. コンテナイメージのプッシュ
```bash
docker push <TODO_REGION>-docker.pkg.dev/<TODO_PROJECT_ID>/<TODO_REPOSITORY>/spacebase:latest
```

## 4. Cloud Run へのデプロイ
```bash
gcloud run deploy spacebase \
  --image=<TODO_REGION>-docker.pkg.dev/<TODO_PROJECT_ID>/<TODO_REPOSITORY>/spacebase:latest \
  --platform=managed \
  --region=<TODO_CLOUD_RUN_REGION> \
  --allow-unauthenticated \
  --set-env-vars GO_ENV=<TODO_ENV>,REDIS_HOST=<TODO_REDIS_HOST>,REDIS_PORT=<TODO_REDIS_PORT>,APP_PORT=<TODO_APP_PORT>,REDIS_TTL=<TODO_REDIS_TTL>,VALKEY=<TODO_VALKEY>
```

## 5. 動作検証
```bash
URL=$(gcloud run services describe spacebase \
  --platform=managed \
  --region=<TODO_CLOUD_RUN_REGION> \
  --format="value(status.url)")
echo "アクセスURL: $URL"
```
- Cloud Run コンソールまたは以下のコマンドでサービスの URL を取得し、ブラウザで動作確認します:

  gcloud run services describe spacebase --platform=managed --region=asia-northeast1 --format="value(status.url)"

## デプロイ時の設定
Docker イメージを GCP にデプロイする際は、以下の環境変数を使用して設定を渡します。これらは `--set-env-vars` オプションで指定され、アプリ起動時に `config` パッケージで読み込まれます。

### Secret Manager からの設定取得
Cloud Run 上で Secret Manager を利用する場合、環境変数 `SECRET_NAME` に Secret のリソース名を指定し、アプリ起動時に Secret Manager から一括で設定をロードします。  
Example:
```bash
export SECRET_NAME="projects/<PROJECT_ID>/secrets/<SECRET_ID>/versions/latest"
gcloud run deploy spacebase \
  --image=<IMAGE_URL> \
  --platform=managed \
  --region=<TODO_CLOUD_RUN_REGION> \
  --allow-unauthenticated \
  --set-env-vars SECRET_NAME=$SECRET_NAME
```
Secret Manager に格納する JSON は以下のような構造です:
```json
{
  "redis_host": "127.0.0.1",
  "redis_port": "6379",
  "app_port": "8080",
  "redis_ttl": "24h",
  "valkey": "実際の機密鍵"
}
```

```bash
export GO_ENV=prod                 # config/config.prod.yaml を利用
export REDIS_HOST="実際のRedisホスト"
export REDIS_PORT="実際のRedisポート"
export APP_PORT="実際のアプリ起動ポート"
export REDIS_TTL="データ有効期限(例:24h)"
export VALKEY="実際の機密鍵"