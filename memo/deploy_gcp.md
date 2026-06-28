作成した Docker イメージを GCP にデプロイするための具体的な手順を説明します。

本プロジェクトには、すでに Cloud Run および Artifact Registry を使用したデプロイ手順書 `docs/DEPLOYMENT.md` と、自動デプロイ用スクリプト `deploy.sh` が用意されています。

以下にデプロイの全体像と具体的なステップをまとめました。

---

### 1. 概要
本プロジェクトのアプリケーションは、**Artifact Registry** にコンテナイメージを格納し、サーバーレス実行環境である **Cloud Run** にデプロイする構成を想定しています。
あらかじめ用意されている `deploy.sh` スクリプトのプレースホルダーを書き換えて実行するか、手動でコマンドを順に実行することでデプロイが完了します。

---

### 2. 主なファイル
- `docs/DEPLOYMENT.md`: Cloud Run / Secret Manager へのデプロイ手順の詳細なドキュメント。
- `deploy.sh`: ビルド、プッシュ、Cloud Run デプロイを一連の流れで行うシェルスクリプト。

---

### 3. デプロイ手順（5つのステップ）

#### ステップ 1: gcloud CLI の準備と認証
ローカル環境から GCP を操作するために、Google Cloud SDK (`gcloud`) をセットアップし、Docker 認証を行います。

```bash
# GCPへのログイン
gcloud auth login

# 対象のプロジェクトを設定
gcloud config set project <YOUR_PROJECT_ID>

# Artifact Registry への Docker 認証を設定 (例: asia-northeast1 の場合)
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
```

#### ステップ 2: Artifact Registry でリポジトリを作成
コンテナイメージを格納するリポジトリを作成します。
GCPコンソール、または以下のコマンドで作成できます。

```bash
gcloud artifacts repositories create <YOUR_REPOSITORY_NAME> \
    --repository-format=docker \
    --location=asia-northeast1 \
    --description="Spacebase docker repository"
```

#### ステップ 3: 設定値の準備
プロジェクトのルートにある `deploy.sh` のプレースホルダー（`<TODO_...>`）を実際の GCP 環境の値に書き換えます。

```bash
# deploy.sh 内の編集箇所
REGION="asia-northeast1"               # Artifact Registry のリージョン
PROJECT_ID="<YOUR_PROJECT_ID>"        # GCPプロジェクトID
REPOSITORY="<YOUR_REPOSITORY_NAME>"    # ステップ2で作成したリポジトリ名
CLOUD_RUN_REGION="asia-northeast1"    # Cloud Run をデプロイするリージョン
```

また、必要な環境変数をローカルでエクスポートしておきます。
```bash
export GO_ENV=prod
export REDIS_HOST="<YOUR_REDIS_HOST>"
export REDIS_PORT="6379"
export APP_PORT="8080"
export REDIS_TTL="24h"
export VALKEY="<YOUR_SECRET_VALKEY>"
```

#### ステップ 4: イメージのビルドとプッシュ
手動で行う場合は以下のコマンドを実行します（`deploy.sh` を実行する場合は自動で行われます）。

```bash
# イメージのビルド
docker build -t asia-northeast1-docker.pkg.dev/<YOUR_PROJECT_ID>/<YOUR_REPOSITORY_NAME>/spacebase:latest .

# イメージのプッシュ
docker push asia-northeast1-docker.pkg.dev/<YOUR_PROJECT_ID>/<YOUR_REPOSITORY_NAME>/spacebase:latest
```

#### ステップ 5: Cloud Run へのデプロイ
コンテナイメージから Cloud Run サービスを起動します。

```bash
gcloud run deploy spacebase \
  --image=asia-northeast1-docker.pkg.dev/<YOUR_PROJECT_ID>/<YOUR_REPOSITORY_NAME>/spacebase:latest \
  --platform=managed \
  --region=asia-northeast1 \
  --allow-unauthenticated \
  --set-env-vars GO_ENV=$GO_ENV,REDIS_HOST=$REDIS_HOST,REDIS_PORT=$REDIS_PORT,APP_PORT=$APP_PORT,REDIS_TTL=$REDIS_TTL,VALKEY=$VALKEY
```

デプロイ完了後、ターミナルに表示される **Service URL** にアクセスして動作確認を行います。

---

### 4. 技術的な注意事項（セキュリティと管理）
* **機密情報の保護**: `VALKEY` やデータベース接続情報などの機密情報は、直接環境変数（`--set-env-vars`）で渡すのではなく、**Secret Manager** に格納して Cloud Run からマウントまたは参照することを強く推奨します。具体的な方法は `docs/DEPLOYMENT.md` の「Secret Manager からの設定取得」セクションを参照してください。
* **IAM 権限**: デプロイを実行するユーザー（またはサービスアカウント）には、以下の権限が必要です。
  - `Artifact Registry 書き込み` (Artifact Registry ライター)
  - `Cloud Run 管理者` (Cloud Run 管理者)
  - `サービス アカウント ユーザー`

---

### 5. 完了条件
* `docker build` および `docker push` が正常に終了し、Artifact Registry にイメージが登録されている。
* `gcloud run deploy` が成功し、出力された URL からアプリケーションが正常に応答する。

---

上記の手順や、プロジェクト内の `deploy.sh`, `docs/DEPLOYMENT.md` についてご不明な点や、さらにカスタマイズしたい部分はありますでしょうか？
もし、デプロイスクリプトの具体的な書き換えや、Secret Manager 連携用の設定変更などの具体的な作業が必要であれば、**Act mode に切り替えて**（画面下のボタンより）ご指示ください。どのようなご質問でもお気軽にどうぞ！