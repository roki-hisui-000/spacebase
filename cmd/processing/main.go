package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
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
	// gRPC SpaceService へ接続して GRPCSpace アダプタを初期化
	spaceAddr := os.Getenv("SPACE_SERVICE_ADDR")
	if spaceAddr == "" {
		spaceAddr = "localhost:50052"
	}

	log.Printf("Connecting to middleware SpaceService at %s", spaceAddr)
	conn, err := grpc.Dial(spaceAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("failed to connect to middleware SpaceService: %v", err)
	}
	defer conn.Close()

	sp := space.NewGRPCSpace(conn)

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

