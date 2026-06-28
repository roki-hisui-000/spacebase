package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/roki-hisui/work/spacebase/config"
	"github.com/roki-hisui/work/spacebase/pkg/model"

	"github.com/redis/go-redis/v9"
)

type ProfileService struct {
	rdb *redis.Client
}

func NewProfileService(rdb *redis.Client) *ProfileService {
	return &ProfileService{
		rdb: rdb,
	}
}

func (s *ProfileService) SaveProfile(ctx context.Context, profile *model.UserProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("user:%s:profile", profile.ID)
	err = s.rdb.Set(ctx, key, data, config.RedisTTL()).Err()
	if err != nil {
		fmt.Printf("Error saving profile to Redis (key: %s): %v\n", key, err)
		return err
	}
	return nil
}

func (s *ProfileService) GetProfile(ctx context.Context, id string) (*model.UserProfile, error) {
	key := fmt.Sprintf("user:%s:profile", id)
	data, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var profile model.UserProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
