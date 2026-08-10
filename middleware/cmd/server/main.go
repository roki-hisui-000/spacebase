package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/roki-hisui/work/spacebase/config"
	"github.com/roki-hisui/work/spacebase/internal/space"
	"github.com/roki-hisui/work/spacebase/middleware/internal/gateway"
)

func main() {
	// RedisSpace アダプタを初期化 (Valkey 互換)

	sp := space.NewRedisSpace(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost(), config.RedisPort()),
	})

	// Gateway をモードに応じて生成 (sync/async)
	gw, err := gateway.NewGateway(sp)
	if err != nil {
		log.Fatalf("Gateway 初期化失敗: %v", err)
	}

	// HTTP サーバーを Gateway ハンドラで起動
	server := &http.Server{
		Handler:      gw,
		Addr:         ":" + config.AppPort(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("仮想ミドルウェアサーバーを %s で起動中", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
