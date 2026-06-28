package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// 1. 保存したいデータの構造体を定義（JSONタグをつける）
type UserProfile struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Languages []string `json:"languages"`
}

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

	// テスト用のデータ（Processing Unitが処理するデータオブジェクト）
	user := UserProfile{
		ID:        100,
		Name:      "Roki",
		Status:    "active",
		Languages: []string{"Go", "Java"},
	}

	// ==========================================
	// パターンA: JSONに変換して Valkey に【保存】
	// ==========================================

	// 構造体を JSON（[]byte）に変換
	jsonData, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("JSON変換に失敗: %v", err)
	}

	// キー「user:100:profile」としてJSONデータを保存（有効期限は24時間と仮定）
	err = rdb.Set(ctx, "user:100:profile", jsonData, 24*time.Hour).Err()
	if err != nil {
		log.Fatalf("Valkeyへの保存に失敗: %v", err)
	}
	fmt.Println("JSONデータをValkeyに保存しました。")

	// ==========================================
	// パターンB: Valkey から読み込んで構造体に【復元】
	// ==========================================

	// ValkeyからJSON文字列を取得
	savedJson, err := rdb.Get(ctx, "user:100:profile").Result()
	if err != nil {
		log.Fatalf("Valkeyからの読み込みに失敗: %v", err)
	}

	// 復元先の空の構造体を用意
	var fetchedUser UserProfile

	// JSON文字列を構造体にマッピング（デシリアライズ）
	err = json.Unmarshal([]byte(savedJson), &fetchedUser)
	if err != nil {
		log.Fatalf("JSONから構造体への復元に失敗: %v", err)
	}

	// 復元されたデータの確認
	fmt.Printf("復元されたユーザー名: %s\n", fetchedUser.Name)
	fmt.Printf("得意言語の1つ目: %s\n", fetchedUser.Languages[0])
}
