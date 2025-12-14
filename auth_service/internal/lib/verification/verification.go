package verification

import (
	"auth_service/internal/models"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Publisher interface {
	SendMessage(ctx context.Context, msg models.Message) error
}

func VerifyUserEmail(
	ctx context.Context,
	log *slog.Logger,
	pub Publisher,
	tokenTTL time.Duration,
	tokenSecret string,
	userID int64,
	url, email string,
) error {
	token, err := generateVerificationToken(userID, tokenTTL, tokenSecret)
	if err != nil {
		log.Error("failed to generate token", slog.Any("err", err))

		return err
	}

	verifyLink := fmt.Sprintf("%s/verify?token=%s", url, token)

	msg := models.Message{
		Email: email,
		Link:  verifyLink,
	}

	if err := pub.SendMessage(ctx, msg); err != nil {
		log.Error("failed to send verification link", slog.Any("err", err))
	}

	return nil
}

func ParseVerificationToken(tokenStr, secret string) (int64, error) {
	const op = "verification.ParseVerificationToken"

	claims := jwt.MapClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method", op)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("%s: failed to parse token: %w", op, err)
	}

	if !parsedToken.Valid {
		return 0, fmt.Errorf("%s: invalid token", op)
	}

	if purpose, ok := claims["purpose"].(string); !ok || purpose != "email_verification" {
		return 0, fmt.Errorf("%s: invalid token purpose", op)
	}

	if expFloat, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(expFloat) {
			return 0, fmt.Errorf("%s: token expired", op)
		}
	} else {
		return 0, fmt.Errorf("%s: missing exp claim", op)
	}

	subFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, fmt.Errorf("%s: missing sub claim", op)
	}

	return int64(subFloat), nil
}

func generateVerificationToken(userID int64, tokenTTL time.Duration, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID,
		"purpose": "email_verification",
		"exp":     time.Now().Add(tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
