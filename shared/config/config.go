package config

import (
	"context"
	"encoding/json"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Config holds application settings loaded from Secret Manager or environment variables.
type Config struct {
	RedisHost string `json:"redis_host"`
	RedisPort string `json:"redis_port"`
	AppPort   string `json:"app_port"`
	RedisTTL  string `json:"redis_ttl"`
	ValKey    string `json:"valkey"`
}

var cfg Config

func init() {
	// SECRET_NAME 環境変数が設定されていれば Secret Manager から設定をロード
	if name := os.Getenv("SECRET_NAME"); name != "" {
		if err := LoadFromSecret(context.Background(), name); err != nil {
			panic("failed to load config from Secret Manager: " + err.Error())
		}
	}
}

// LoadFromSecret fetches and unmarshals JSON-config from Secret Manager.
func LoadFromSecret(ctx context.Context, secretName string) error {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return err
	}
	req := &smpb.AccessSecretVersionRequest{Name: secretName}
	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.Payload.Data, &cfg)
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}

// RedisHost returns Redis host.
func RedisHost() string {
	if cfg.RedisHost != "" {
		return cfg.RedisHost
	}
	return getEnv("REDIS_HOST", "127.0.0.1")
}

// RedisPort returns Redis port.
func RedisPort() string {
	if cfg.RedisPort != "" {
		return cfg.RedisPort
	}
	return getEnv("REDIS_PORT", "6379")
}

// AppPort returns HTTP server port.
func AppPort() string {
	if cfg.AppPort != "" {
		return cfg.AppPort
	}
	return getEnv("APP_PORT", "8080")
}

// RedisTTL returns Redis TTL duration.
func RedisTTL() time.Duration {
	var str string
	if cfg.RedisTTL != "" {
		str = cfg.RedisTTL
	} else {
		str = getEnv("REDIS_TTL", "24h")
	}
	d, err := time.ParseDuration(str)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// ValKey returns the ValKey.
func ValKey() string {
	if cfg.ValKey != "" {
		return cfg.ValKey
	}
	return getEnv("VALKEY", "")
}
