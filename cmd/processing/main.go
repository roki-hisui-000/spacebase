package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/roki-hisui/work/spacebase/config"
	processingpb "github.com/roki-hisui/work/spacebase/internal/processing"
	"github.com/roki-hisui/work/spacebase/internal/space"
	"github.com/roki-hisui/work/spacebase/pkg/model"
	"google.golang.org/grpc"
)

// processingServer implements the ProcessingUnit gRPC service.
type processingServer struct {
	processingpb.UnimplementedProcessingUnitServer
	space space.Space
}

// RegisterUser generates a new UUID, stores the profile, and returns a response.
func (s *processingServer) RegisterUser(ctx context.Context, req *processingpb.RegisterUserRequest) (*processingpb.RegisterUserResponse, error) {
	// Generate server-side UUID
	id := uuid.New().String()
	key := fmt.Sprintf("user:%s:profile", id)

	// すべての情報を含めたUserProfile構造体を組み立て
	profile := model.UserProfile{
		ID:        id,
		Email:     req.GetName(),
		Status:    req.GetStatus(),
		Languages: req.GetLanguages(),
	}

	// JSONにシリアライズ
	value, err := json.Marshal(profile)
	if err != nil {
		return &processingpb.RegisterUserResponse{
			Success: false,
			Message: "failed to serialize profile: " + err.Error(),
		}, nil
	}

	ttl := config.RedisTTL()

	tuple := space.Tuple{
		Key:   key,
		Value: value,
		TTL:   ttl,
	}
	if err := s.space.Put(ctx, tuple); err != nil {
		return &processingpb.RegisterUserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &processingpb.RegisterUserResponse{
		Success: true,
		Message: "User registered successfully",
	}, nil
}

func main() {
	// RedisSpace adapter initialization (using Redis protocol client for Valkey compatibility)
	sp := space.NewRedisSpace(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost(), config.RedisPort()),
	})

	// Start gRPC server
	addr := os.Getenv("PROCESSING_UNIT_ADDR")
	if addr == "" {
		addr = ":50051"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer()
	processingpb.RegisterProcessingUnitServer(grpcServer, &processingServer{space: sp})

	log.Printf("ProcessingUnit gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC serve failed: %v", err)
	}
}
