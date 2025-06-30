package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"microservices/auth-service/application/usecases"
	"microservices/auth-service/domain/entities"
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

// Implementasi UserRepository
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
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret")

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
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret")

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
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret")

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
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret")

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
	authUC := usecases.NewAuthUseCase(mockUserRepo, mockTokenRepo, "test-secret")

	// Mock data
	refreshToken := "valid.refresh.token"
	tokenID := "token-id-123"
	userID := "user-123"

	// Mock expectations
	mockTokenRepo.On("IsTokenRevoked", mock.Anything, tokenID).Return(false)
	mockTokenRepo.On("GetUserIDByTokenID", mock.Anything, tokenID).Return(userID, nil)

	user := &entities.User{
		ID:    userID,
		Email: "user@example.com",
		Role:  entities.ClientRole,
	}
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	mockTokenRepo.On("RevokeToken", mock.Anything, tokenID).Return(nil)
	mockTokenRepo.On("StoreToken", mock.Anything, mock.Anything, userID).Return(nil)

	// Execute
	newAccess, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
