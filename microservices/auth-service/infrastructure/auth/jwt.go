package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenExpiry = 7 * 24 * time.Hour
)

type JWTAuth struct {
	secret string
}

func NewJWTAuth(secret string) *JWTAuth {
	return &JWTAuth{secret: secret}
}

type CustomClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (ja *JWTAuth) GenerateAccessToken(userID, role string) (string, error) {
	accessClaims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiry)),
			ID:        uuid.NewString(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessToken.SignedString([]byte(ja.secret))
}

func (ja *JWTAuth) GenerateTokens(userID, role string) (string, string, error) {
	accessToken, err := ja.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := ja.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (ja *JWTAuth) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ja.secret), nil
	})

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// Add refresh token handling
func (ja *JWTAuth) GenerateRefreshToken() (string, error) {
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiry)),
		ID:        uuid.NewString(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	return refreshToken.SignedString([]byte(ja.secret))
}

func (ja *JWTAuth) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ja.secret), nil
	})

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func (ja *JWTAuth) RotateTokens(userID, role string) (string, string, error) {
	return ja.GenerateTokens(userID, role)
}
