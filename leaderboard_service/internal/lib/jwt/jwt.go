package jwt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingAuthHeader    = errors.New("missing Authorization header")
	ErrInvalidAuthHeader    = errors.New("invalid Authorization header")
	ErrInvalidToken         = errors.New("invalid token")
	ErrMissingUserIDClaim   = errors.New("user_id missing in token")
	ErrMissingUsernameClaim = errors.New("username missing in token")
)

type JWTParser struct {
	Secret string
}

func New(secret string) *JWTParser {
	return &JWTParser{
		Secret: secret,
	}
}

// * ParseToken извлекает userID и username из JWT токена
func (p *JWTParser) ParseToken(authHeader string) (int64, string, error) {
	if authHeader == "" {
		return 0, "", ErrMissingAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return 0, "", ErrInvalidAuthHeader
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", ErrInvalidToken
	}

	userID, ok := claims["uid"].(int64)
	if !ok {
		return 0, "", ErrMissingUserIDClaim
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return 0, "", ErrMissingUsernameClaim
	}

	return userID, username, nil
}
