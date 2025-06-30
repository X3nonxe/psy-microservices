package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"microservices/auth-service/application/usecases"
	"microservices/auth-service/domain/entities"
	"microservices/auth-service/infrastructure/auth"
)

type MockUserRepository struct {
	mock.Mock
}

type MockTokenRepository struct {
	mock.Mock
}

// Implementasi TokenRepository
func (m *MockTokenRepository) StoreToken(ctx context.Context, tokenID, userID string) error {
	args := m.Called(ctx, tokenID, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) IsTokenRevoked(ctx context.Context, tokenID string) bool {
	args := m.Called(ctx, tokenID)
	return args.Bool(0)
}

func (m *MockTokenRepository) RevokeToken(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockTokenRepository) GetUserIDByTokenID(ctx context.Context, tokenID string) (string, error) {
	args := m.Called(ctx, tokenID)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Mock expectations
	mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return((*entities.User)(nil), nil)
	mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)

	// Execute
	user, err := authUC.Register(context.Background(), "test@example.com", "password123", entities.ClientRole)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, entities.ClientRole, user.Role)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Mock: User exists but wrong password
	existingUser := &entities.User{
		ID:           "user-123",
		Email:        "user@example.com",
		PasswordHash: "correct-hash", // In real test, use valid Argon2 hash
		Role:         entities.ClientRole,
		CreatedAt:    time.Now(),
	}
	mockUserRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(existingUser, nil)

	// Execute
	_, _, err := authUC.Login(context.Background(), "user@example.com", "wrong-password")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entities.ErrInvalidCredentials, err)
}

func TestAuthUseCase_Register_EmailExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	existingUser := &entities.User{
		ID:    "user-123",
		Email: "existing@example.com",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	_, err := authUC.Register(context.Background(), "existing@example.com", "password123", entities.ClientRole)

	assert.Error(t, err)
	assert.Equal(t, entities.ErrEmailExists, err)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Gunakan hash password yang valid
	validHash := "$argon2id$v=19$m=65536,t=1,p=4$c2FsdHNhbHRzYWx0$Df7G0CbGqD8N0mF8eH0wZg0bWY0bWY0bWY0bWY0bWY0"
	existingUser := &entities.User{
		ID:           "user-123",
		Email:        "user@example.com",
		PasswordHash: validHash,
		Role:         entities.ClientRole,
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(existingUser, nil)

	_, _, err := authUC.Login(context.Background(), "user@example.com", "wrong-password")

	assert.Error(t, err)
	assert.Equal(t, entities.ErrInvalidCredentials, err)
}

func TestAuthUseCase_RefreshToken_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// User data
	userID := "user-123"
	user := &entities.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  entities.ClientRole,
	}

	// Generate valid refresh token
	jwtAuth := auth.NewJWTAuth("test-secret")
	refreshToken, err := jwtAuth.GenerateRefreshToken()
	assert.NoError(t, err)

	// Parse token to get claims
	claims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)

	// Mock expectations
	mockTokenRepo.On("GetUserIDByTokenID", mock.Anything, claims.ID).Return(userID, nil)
	mockTokenRepo.On("IsTokenRevoked", mock.Anything, claims.ID).Return(false)
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	mockTokenRepo.On("RevokeToken", mock.Anything, claims.ID).Return(nil)

	// Gunakan mock.Anything untuk token ID baru karena nilainya acak
	mockTokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("string"), userID).Return(nil)

	// Execute
	accessToken, newRefreshToken, err := authUC.RefreshToken(context.Background(), refreshToken)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, newRefreshToken)

	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_Revoked(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Generate valid refresh token
	jwtAuth := auth.NewJWTAuth("test-secret")
	refreshToken, err := jwtAuth.GenerateRefreshToken()
	assert.NoError(t, err)

	// Parse token to get claims
	claims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)

	// Mock expectations
	userID := "user-123"

	// Mock GetUserIDByTokenID untuk mengembalikan userID
	mockTokenRepo.On("GetUserIDByTokenID", mock.Anything, claims.ID).Return(userID, nil)

	// Mock IsTokenRevoked untuk mengembalikan true (token dicabut)
	mockTokenRepo.On("IsTokenRevoked", mock.Anything, claims.ID).Return(true)

	// Execute
	_, _, err = authUC.RefreshToken(context.Background(), refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entities.ErrTokenRevoked, err)

	// Pastikan hanya GetUserIDByTokenID dan IsTokenRevoked yang dipanggil
	mockTokenRepo.AssertCalled(t, "GetUserIDByTokenID", mock.Anything, claims.ID)
	mockTokenRepo.AssertCalled(t, "IsTokenRevoked", mock.Anything, claims.ID)

	// Pastikan fungsi lainnya tidak dipanggil
	mockUserRepo.AssertNotCalled(t, "FindByID")
	mockTokenRepo.AssertNotCalled(t, "RevokeToken")
	mockTokenRepo.AssertNotCalled(t, "StoreToken")
}

func TestAuthUseCase_Logout_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Generate valid refresh token
	jwtAuth := auth.NewJWTAuth("test-secret")
	refreshToken, err := jwtAuth.GenerateRefreshToken()
	assert.NoError(t, err)

	// Parse token to get claims
	claims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)

	// Mock expectations
	mockTokenRepo.On("RevokeToken", mock.Anything, claims.ID).Return(nil)

	err = authUC.Logout(context.Background(), refreshToken)

	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_InvalidToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	err := authUC.Logout(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Equal(t, entities.ErrInvalidToken, err)
}

func TestAuthUseCase_RefreshToken_InvalidToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	_, _, err := authUC.RefreshToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Equal(t, entities.ErrInvalidToken, err)
}

func TestAuthUseCase_RefreshToken_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret", nil)

	// Generate valid refresh token
	jwtAuth := auth.NewJWTAuth("test-secret")
	refreshToken, err := jwtAuth.GenerateRefreshToken()
	assert.NoError(t, err)

	// Parse token to get claims
	claims, err := jwtAuth.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)

	// Mock expectations
	mockTokenRepo.On("GetUserIDByTokenID", mock.Anything, claims.ID).Return("invalid-user-id", nil)
	mockTokenRepo.On("IsTokenRevoked", mock.Anything, claims.ID).Return(false)
	mockUserRepo.On("FindByID", mock.Anything, "invalid-user-id").Return(nil, errors.New("not found"))

	_, _, err = authUC.RefreshToken(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Equal(t, entities.ErrUserNotFound, err)
}
