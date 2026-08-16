package service

import (
	"context"
	"time"

	processingpb "github.com/roki-hisui/work/spacebase/internal/processing"
	"github.com/roki-hisui/work/spacebase/internal/space"
)

// SpaceServiceServer implements SpaceService gRPC service.
type SpaceServiceServer struct {
	processingpb.UnimplementedSpaceServiceServer
	space space.Space
}

// NewSpaceServiceServer creates a new SpaceServiceServer.
func NewSpaceServiceServer(sp space.Space) *SpaceServiceServer {
	return &SpaceServiceServer{space: sp}
}

// Put handles tuple insertion.
func (s *SpaceServiceServer) Put(ctx context.Context, req *processingpb.PutRequest) (*processingpb.PutResponse, error) {
	if req.Tuple == nil {
		return &processingpb.PutResponse{Success: false, Error: "tuple is nil"}, nil
	}
	tuple := space.Tuple{
		Key:   req.Tuple.Key,
		Value: req.Tuple.Value,
		TTL:   time.Duration(req.Tuple.TtlMs) * time.Millisecond,
	}
	err := s.space.Put(ctx, tuple)
	if err != nil {
		return &processingpb.PutResponse{Success: false, Error: err.Error()}, nil
	}
	return &processingpb.PutResponse{Success: true}, nil
}

// Get handles tuple retrieval.
func (s *SpaceServiceServer) Get(ctx context.Context, req *processingpb.GetRequest) (*processingpb.GetResponse, error) {
	tuple, err := s.space.Get(ctx, req.Key)
	if err != nil {
		return &processingpb.GetResponse{Found: false, Error: err.Error()}, nil
	}
	return &processingpb.GetResponse{
		Tuple: &processingpb.TupleMessage{
			Key:   tuple.Key,
			Value: tuple.Value,
			TtlMs: int64(tuple.TTL / time.Millisecond),
		},
		Found: true,
	}, nil
}

// Keys handles pattern matching retrieval.
func (s *SpaceServiceServer) Keys(ctx context.Context, req *processingpb.KeysRequest) (*processingpb.KeysResponse, error) {
	keys, err := s.space.Keys(ctx, req.Pattern)
	if err != nil {
		return &processingpb.KeysResponse{Error: err.Error()}, nil
	}
	return &processingpb.KeysResponse{Keys: keys}, nil
}
