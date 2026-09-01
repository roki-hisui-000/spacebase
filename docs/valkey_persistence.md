# Valkey 永続化設計・設定ドキュメント

本ドキュメントでは、ローカル開発環境およびコンテナ環境（Docker Compose 等）における Valkey（インメモリデータストア）のデータ永続化設計について説明します。

## 1. 永続化の仕組み
Valkey のデータはデフォルトでメモリ上のみに存在し、コンテナを破棄（`docker compose down`）すると消失してしまいます。このため、以下の2つの方法を併用してデータを物理ディスクに永続化しています。

1. **AOF (Append Only File) [推奨]**
   - コマンドログをファイルに逐一追記していく方式。データの欠損が極めて少なく、信頼性が高い。
2. **RDB (Redis Database) スナップショット**
   - 指定された秒数とキーの変更件数をトリガーに、メモリのスナップショットをディスクに書き出す。

Docker Compose 環境では、コンテナ起動時に以下のコマンドオプションを指定して永続化を有効化しています。
```bash
valkey-server --save 60 1 --appendonly yes
```
* `--save 60 1`: 「60秒間に1件以上のキー変更があった場合にRDBスナップショットを書き出す」設定。
* `--appendonly yes`: AOFを有効化。

---

## 2. ディレクトリマウント設計
コンテナ内の永続化データディレクトリ `/data` を、ホストマシン上の物理ディレクトリにマウントします。

### 2.1 環境変数（プロパティ）によるマウントパスの動的設定
マウントパスは、環境変数 `VALKEY_DATA_DIR` を通じて動的に変更できるようになっています。
`docker-compose.yml` での定義は以下のようになっています。
```yaml
  valkey:
    ...
    volumes:
      - ${VALKEY_DATA_DIR:-./data/valkey}:/data
```
- ホスト側で `VALKEY_DATA_DIR` が設定されていない場合、デフォルト値として `./data/valkey`（リポジトリルートからの相対パス）が利用されます。
- この `./data` ディレクトリはローカル開発用であるため、誤ってコミットされないように `.gitignore` で追跡対象外（`/data/`）に指定しています。

---

## 3. 上位環境・他環境への適用方法

今後、ローカル以外の別の環境（ステージング、本番など）で Valkey をコンテナ稼働させる場合、用途に合わせてマウント先フォルダを変更できます。

### 3.1 別の Docker Compose 環境（VM、オンプレ等）での実行時
VM などのホスト上で実行する場合、`/var/lib/valkey` などシステム管理用の適切な永続化ディレクトリに変更したいケースがあります。この場合、以下のいずれかの方法でディレクトリを変更します。

#### 方法 A: `.env` ファイルでの上書き設定
デプロイ環境のプロジェクトルートに配置する `.env` ファイルに設定を追加します。
```ini
# 上位環境用に永続化先ディレクトリを指定
VALKEY_DATA_DIR=/var/lib/valkey-data
```

#### 方法 B: 環境変数経由での設定
コンテナ起動時に、ホストの環境変数を渡して実行します。
```bash
export VALKEY_DATA_DIR="/mnt/persistent-volume/valkey"
make up
```

### 3.2 Kubernetes (GKE等) への移行時
本番環境で Kubernetes を用いる場合、Docker Compose のローカルボリュームマウントではなく、PV（PersistentVolume）および PVC（PersistentVolumeClaim）を使用してディスクをマウントします。

**PVC 定義例 (`valkey-pvc.yaml`):**
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: valkey-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

**Deployment / StatefulSet 定義例:**
```yaml
spec:
  containers:
    - name: valkey
      image: valkey/valkey:8
      command: ["valkey-server", "--save", "60", "1", "--appendonly", "yes"]
      volumeMounts:
        - name: valkey-data
          mountPath: /data
  volumes:
    - name: valkey-data
      persistentVolumeClaim:
        claimName: valkey-pvc
```
このようにすることで、コンテナの再起動やノードの再スケジュール時にも、クラウドプロバイダ（GCPのPersistent Disk等）でプロビジョニングされた永続データが自動的に維持されます。
