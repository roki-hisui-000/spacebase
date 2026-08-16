package space

import (
	"context"
	"fmt"
	"time"

	processingpb "github.com/roki-hisui/work/spacebase/internal/processing"
	"google.golang.org/grpc"
)

// GRPCSpace is an implementation of Space backed by SpaceService gRPC.
type GRPCSpace struct {
	client processingpb.SpaceServiceClient
}

// NewGRPCSpace creates a new GRPCSpace.
func NewGRPCSpace(conn *grpc.ClientConn) *GRPCSpace {
	return &GRPCSpace{
		client: processingpb.NewSpaceServiceClient(conn),
	}
}

// Put stores a tuple via SpaceService.
func (g *GRPCSpace) Put(ctx context.Context, tuple Tuple) error {
	req := &processingpb.PutRequest{
		Tuple: &processingpb.TupleMessage{
			Key:   tuple.Key,
			Value: tuple.Value,
			TtlMs: int64(tuple.TTL / time.Millisecond),
		},
	}
	resp, err := g.client.Put(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("gRPC put failed: %s", resp.Error)
	}
	return nil
}

// Get retrieves a tuple by key via SpaceService.
func (g *GRPCSpace) Get(ctx context.Context, key string) (Tuple, error) {
	req := &processingpb.GetRequest{Key: key}
	resp, err := g.client.Get(ctx, req)
	if err != nil {
		return Tuple{}, err
	}
	if resp.Error != "" {
		return Tuple{}, fmt.Errorf("gRPC get failed: %s", resp.Error)
	}
	if !resp.Found {
		return Tuple{}, fmt.Errorf("key not found: %s", key)
	}
	return Tuple{
		Key:   resp.Tuple.Key,
		Value: resp.Tuple.Value,
		TTL:   time.Duration(resp.Tuple.TtlMs) * time.Millisecond,
	}, nil
}

// Keys returns keys matching pattern via SpaceService.
func (g *GRPCSpace) Keys(ctx context.Context, pattern string) ([]string, error) {
	req := &processingpb.KeysRequest{Pattern: pattern}
	resp, err := g.client.Keys(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("gRPC keys failed: %s", resp.Error)
	}
	return resp.Keys, nil
}
