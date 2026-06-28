package space

import (
	"context"
	"time"
)

// Tuple represents a key-value entry with TTL.
type Tuple struct {
	Key   string
	Value []byte
	TTL   time.Duration
}

// Space defines abstract operations for tuple space or key-value grid.
type Space interface {
	// Put stores a tuple into the space.
	Put(ctx context.Context, tuple Tuple) error
	// Get retrieves a tuple by key.
	Get(ctx context.Context, key string) (Tuple, error)
	// Keys returns all keys matching a pattern.
	Keys(ctx context.Context, pattern string) ([]string, error)
}
