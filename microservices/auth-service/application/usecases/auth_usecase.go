package usecases

import (
	"context"
	"log"
	"microservices/auth-service/domain/entities"
	"microservices/auth-service/domain/repositories"
	"microservices/auth-service/infrastructure/auth"
	"time"

	"go.uber.org/zap"
)

type AuthUseCase struct {
	userRepo  repositories.UserRepository
	tokenRepo repositories.TokenRepository
	jwtAuth   *auth.JWTAuth
	logger    *zap.Logger
}
func NewAuthUseCase(userRepo repositories.UserRepository, tokenRepo repositories.TokenRepository, jwtSecret string, logger *zap.Logger) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtAuth:   auth.NewJWTAuth(jwtSecret),
		logger:    logger,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password string, role entities.Role) (*entities.User, error) {
	// Validasi email unik
	if existing, _ := uc.userRepo.FindByEmail(ctx, email); existing != nil {
		return nil, entities.ErrEmailExists
	}

	// Hash password
	hashedPassword, err := auth.Argon2Hash(password)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		ID:           auth.GenerateUUID(),
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	if err := uc.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", entities.ErrInvalidCredentials
	}

	if !auth.Argon2Verify(password, user.PasswordHash) {
		return "", "", entities.ErrInvalidCredentials
	}

	accessToken, refreshToken, err := uc.jwtAuth.GenerateTokens(user.ID, string(user.Role))
	if err != nil {
		return "", "", err
	}

	// Simpan refresh token ke repository
	refreshClaims, err := uc.jwtAuth.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("failed to validate refresh token: %v", err)
	} else {
		if err := uc.tokenRepo.StoreToken(ctx, refreshClaims.ID, user.ID); err != nil {
			log.Printf("failed to store refresh token: %v", err)
		}
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := uc.jwtAuth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", entities.ErrInvalidToken
	}

	// Dapatkan userID dari token repository
	userID, err := uc.tokenRepo.GetUserIDByTokenID(ctx, claims.ID)
	if err != nil {
		return "", "", entities.ErrInvalidToken
	}

	// Periksa apakah token revoked
	if uc.tokenRepo.IsTokenRevoked(ctx, claims.ID) {
		return "", "", entities.ErrTokenRevoked
	}

	// Dapatkan user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", entities.ErrUserNotFound
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := uc.jwtAuth.RotateTokens(user.ID, string(user.Role))
	if err != nil {
		return "", "", err
	}

	// Parse new refresh token claims
	newRefreshClaims, err := uc.jwtAuth.ValidateRefreshToken(newRefreshToken)
	if err != nil {
		return "", "", err
	}

	// Revoke old refresh token
	if err := uc.tokenRepo.RevokeToken(ctx, claims.ID); err != nil {
		uc.logger.Error("failed to revoke token", zap.Error(err))
	}

	// Store new refresh token
	if err := uc.tokenRepo.StoreToken(ctx, newRefreshClaims.ID, user.ID); err != nil {
		uc.logger.Error("failed to store new refresh token", zap.Error(err))
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	claims, err := uc.jwtAuth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return entities.ErrInvalidToken
	}

	return uc.tokenRepo.RevokeToken(ctx, claims.ID)
}
