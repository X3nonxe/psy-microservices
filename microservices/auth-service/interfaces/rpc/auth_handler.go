package rpc

import (
	"context"
	"microservices/auth-service/application/usecases"
	"microservices/auth-service/domain/entities"
	v1 "microservices/auth-service/gen/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	v1.UnimplementedAuthServiceServer
	authUC *usecases.AuthUseCase
}

func NewAuthHandler(authUC *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	user, err := h.authUC.Register(ctx, req.Email, req.Password, entities.Role(req.Role))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}
	return &v1.RegisterResponse{UserId: user.ID}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	accessToken, refreshToken, err := h.authUC.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}
	return &v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
