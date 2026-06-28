# 設定管理の代替案

現在の「YAMLファイル＋環境変数」方式は設定項目が増えると煩雑になります。以下のような手法を利用すると、よりシンプルかつ安全に運用できます。

## 1. GCP Secret Manager から一括取得
- 複数の設定値（REDIS_HOST、REDIS_PORT、APP_PORT、REDIS_TTL、VALKEY など）を JSON 形式で一つの Secret にまとめる  
- 起動時に GCP SDK で Secret をフェッチし、`json.Unmarshal` で構造体に展開  
- 環境変数は Secret 名だけ渡せばよく、実際の値は Secret Manager 管理下

```go
import (
  "context"
  secretmanager "cloud.google.com/go/secretmanager/apiv1"
  "encoding/json"
  smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Config struct {
  RedisHost string `json:"redis_host"`
  RedisPort int    `json:"redis_port"`
  AppPort   int    `json:"app_port"`
  RedisTTL  string `json:"redis_ttl"`
  ValKey    string `json:"valkey"`
}

func LoadFromSecret(ctx context.Context, secretName string) (*Config, error) {
  client, _ := secretmanager.NewClient(ctx)
  req := &smpb.AccessSecretVersionRequest{ Name: secretName }
  resp, err := client.AccessSecretVersion(ctx, req)
  if err != nil {
    return nil, err
  }
  var cfg Config
  if err := json.Unmarshal(resp.Payload.Data, &cfg); err != nil {
    return nil, err
  }
  return &cfg, nil
}
```

## 2. 環境変数に JSON 一括設定
- `--set-env-vars CONFIG_JSON='<JSON文字列>'` のように一つの環境変数で全設定を渡す  
- 起動時に `os.Getenv("CONFIG_JSON")` を `json.Unmarshal` で展開  
- 環境変数数は1つだけで済む

## 3. HashiCorp Vault／AWS SSM Parameter Store
- 外部の設定ストア（Vault、SSM）を利用し、複数のキーを一括取得  
- 同様に JSON 形式や Key-Value でフェッチ可能  

---

以上のいずれかを選択すると、設定ファイル管理や複数環境変数の煩雑さを軽減できます。