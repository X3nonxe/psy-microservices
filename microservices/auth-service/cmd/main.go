package main

import (
	"database/sql"
	"microservices/auth-service/application/usecases"
	"microservices/auth-service/config"
	"microservices/auth-service/infrastructure/logger"
	"microservices/auth-service/infrastructure/persistence"
	"microservices/auth-service/interfaces/middleware"
	"microservices/auth-service/interfaces/rpc"
	"net"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	v1 "microservices/auth-service/gen/auth/v1"
)

func main() {
	cfg := config.Load()
	logger := logger.NewLogger("auth-service")
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	userRepo := persistence.NewPostgresUserRepository(db)
	tokenRepo := persistence.NewRedisTokenRepository(redisClient)
	authUC := usecases.NewAuthUseCase(userRepo, tokenRepo, cfg.JWTSecret, zap.L())

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		zap.L().Fatal("failed to listen", zap.Error(err))
	}

	rateLimiter := middleware.NewRateLimiter(100)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(rateLimiter.UnaryInterceptor()),
		grpc.StreamInterceptor(rateLimiter.StreamInterceptor()),
	)
	v1.RegisterAuthServiceServer(s, rpc.NewAuthHandler(authUC))

	healthHandler := rpc.NewHealthHandler(db, *tokenRepo)
	grpc_health_v1.RegisterHealthServer(s, healthHandler)

	reflection.Register(s)

	zap.L().Info("Starting auth service", zap.String("port", cfg.GRPCPort))
	if err := s.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve gRPC", zap.Error(err))
	}
}
