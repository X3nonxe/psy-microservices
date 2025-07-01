package rpc

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"microservices/auth-service/infrastructure/persistence"
)

type HealthHandler struct {
	db        *sql.DB
	redisRepo persistence.RedisTokenRepository // Tambahkan ini
}

func NewHealthHandler(db *sql.DB, redisRepo persistence.RedisTokenRepository) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redisRepo: redisRepo,
	}
}

func (h *HealthHandler) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	// Check database
	if err := h.db.PingContext(ctx); err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	// Check Redis jika ada
	var zeroRepo persistence.RedisTokenRepository
	if h.redisRepo != zeroRepo {
		if err := h.redisRepo.CheckHealth(ctx); err != nil {
			return &grpc_health_v1.HealthCheckResponse{
				Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			}, nil
		}
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthHandler) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}

func (h *HealthHandler) List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
    // TODO: implement the correct logic for HealthListResponse
    return &grpc_health_v1.HealthListResponse{}, nil
}
