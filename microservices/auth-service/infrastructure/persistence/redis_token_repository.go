package persistence

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisTokenRepository struct {
	client *redis.Client
	prefix string
}

var refreshTokenExpiry = 7 * 24 * time.Hour // Set token expiry to 7 days

func NewRedisTokenRepository(client *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{
		client: client,
		prefix: "refresh_token:",
	}
}

func (r *RedisTokenRepository) StoreToken(ctx context.Context, tokenID, userID string) error {
	expiration := refreshTokenExpiry
	return r.client.Set(ctx, r.prefix+tokenID, userID, expiration).Err()
}

func (r *RedisTokenRepository) IsTokenRevoked(ctx context.Context, tokenID string) bool {
	result, err := r.client.Exists(ctx, r.prefix+tokenID).Result()
	return err != nil || result == 0
}

func (r *RedisTokenRepository) RevokeToken(ctx context.Context, tokenID string) error {
	return r.client.Del(ctx, r.prefix+tokenID).Err()
}

// Fungsi baru untuk mendapatkan UserID berdasarkan TokenID
func (r *RedisTokenRepository) GetUserIDByTokenID(ctx context.Context, tokenID string) (string, error) {
	return r.client.Get(ctx, r.prefix+tokenID).Result()
}

func (r *RedisTokenRepository) CheckHealth(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}
