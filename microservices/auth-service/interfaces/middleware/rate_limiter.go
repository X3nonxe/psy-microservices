package middleware

import (
	"context"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), rps),
	}
}

func (rl *RateLimiter) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !rl.limiter.Allow() {
			return nil, status.Errorf(codes.ResourceExhausted, "too many requests")
		}
		return handler(ctx, req)
	}
}

func (rl *RateLimiter) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !rl.limiter.Allow() {
			return status.Errorf(codes.ResourceExhausted, "too many requests")
		}
		return handler(srv, stream)
	}
}
