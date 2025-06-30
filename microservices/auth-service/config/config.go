package config

import "os"

type Config struct {
	GRPCPort  string
	DBURL     string
	RedisURL  string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		GRPCPort:  getEnv("AUTH_GRPC_PORT", "50051"),
		DBURL:     getEnv("AUTH_DB_URL", "postgres://auth_user:auth_password@postgres:5432/auth_db?sslmode=disable"),
		RedisURL:  getEnv("REDIS_URL", "redis:6379"),
		JWTSecret: getEnv("JWT_SECRET", "default_secret"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
