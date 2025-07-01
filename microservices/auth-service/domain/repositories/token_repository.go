package repositories

import (
	"context"
)

type TokenRepository interface {
	StoreToken(ctx context.Context, tokenID, userID string) error
	IsTokenRevoked(ctx context.Context, tokenID string) bool
	RevokeToken(ctx context.Context, tokenID string) error
	GetUserIDByTokenID(ctx context.Context, tokenID string) (string, error)
}
