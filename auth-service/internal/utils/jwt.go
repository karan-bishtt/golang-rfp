package utils

import (
	"errors"
	"time"

	"github.com/karan-bishtt/auth-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateTokenPair generates both refresh and access tokens
func GenerateTokenPair(userID uint, role string) (refreshToken, accessToken string, err error) {
	cfg := config.Load()

	// Generate access token (15 minutes)
	accessToken, err = generateToken(userID, role, time.Minute*24*1, cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token (7 days)
	refreshToken, err = generateToken(userID, role, time.Hour*24*7, cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

// generateToken creates a JWT token with specified duration
func generateToken(userID uint, role string, duration time.Duration, secret string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Load()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshTokenString string) (newAccessToken string, err error) {
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	cfg := config.Load()

	// Generate new access token
	newAccessToken, err = generateToken(claims.UserID, claims.Role, time.Minute*15, cfg.JWTSecret)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetUserFromToken extracts user information from token
func GetUserFromToken(tokenString string) (userID uint, role string, err error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, "", err
	}

	return claims.UserID, claims.Role, nil
}
