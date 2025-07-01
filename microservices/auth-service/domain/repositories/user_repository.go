package repositories

import (
	"context"
	"microservices/auth-service/domain/entities"
)

// UserRepository : Interface untuk abstract database
type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
}