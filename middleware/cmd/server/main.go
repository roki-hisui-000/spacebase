package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/roki-hisui/work/spacebase/config"
	processingpb "github.com/roki-hisui/work/spacebase/internal/processing"
	"github.com/roki-hisui/work/spacebase/internal/space"
	"github.com/roki-hisui/work/spacebase/middleware/internal/gateway"
	"github.com/roki-hisui/work/spacebase/middleware/internal/service"
	"google.golang.org/grpc"
)

func main() {
	// RedisSpace アダプタを初期化 (Valkey 互換)

	sp := space.NewRedisSpace(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost(), config.RedisPort()),
	})

	// 1. gRPC SpaceService サーバーを起動
	grpcAddr := os.Getenv("SPACE_SERVICE_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50052"
	}
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("gRPC SpaceService listener failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	spaceServer := service.NewSpaceServiceServer(sp)
	processingpb.RegisterSpaceServiceServer(grpcServer, spaceServer)

	go func() {
		log.Printf("ミドルウェア SpaceService gRPC サーバーを %s で起動中", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC SpaceService serve failed: %v", err)
		}
	}()

	// 2. Gateway をモードに応じて生成 (sync/async)
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

