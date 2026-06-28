package space

import (
	"context"
	"os/exec"
	"strings"
)

// ValKeySpace implements Space using ValKey CLI.
type ValKeySpace struct{}

// NewValKeySpace creates a new ValKeySpace.
func NewValKeySpace() *ValKeySpace {
	return &ValKeySpace{}
}

// Put stores a tuple into ValKey, with optional TTL.
func (v *ValKeySpace) Put(ctx context.Context, tuple Tuple) error {
	args := []string{"set", tuple.Key, string(tuple.Value)}
	if tuple.TTL > 0 {
		args = append(args, "-t", tuple.TTL.String())
	}
	cmd := exec.CommandContext(ctx, "valkey-cli", args...)
	return cmd.Run()
}

// Get retrieves a tuple by key from ValKey.
func (v *ValKeySpace) Get(ctx context.Context, key string) (Tuple, error) {
	cmd := exec.CommandContext(ctx, "valkey-cli", "get", key)
	out, err := cmd.Output()
	if err != nil {
		return Tuple{}, err
	}
	return Tuple{Key: key, Value: []byte(strings.TrimSpace(string(out))), TTL: 0}, nil
}

// Keys returns all keys matching the pattern from ValKey.
func (v *ValKeySpace) Keys(ctx context.Context, pattern string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "valkey-cli", "keys", pattern)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}
