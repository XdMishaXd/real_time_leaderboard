package auth

import (
	"auth_service/internal/lib/jwt"
	sl "auth_service/internal/lib/logger"
	"auth_service/internal/lib/verification"
	"auth_service/internal/models"
	"auth_service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrEmailNotVerified   = errors.New("email not verified")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
	refreshTTL  time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, username string, passHash []byte) (uid int64, err error)

	SaveRefreshToken(ctx context.Context, userID int64, appID int32, tokenHash []byte, expiresAt time.Time) error
	UpdateRefreshToken(ctx context.Context, userID int64, oldTokenHash, newTokenHash []byte, expiresAt time.Time) error
	DeleteRefreshToken(ctx context.Context, tokenHash []byte) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	UserByID(ctx context.Context, id int64) (models.User, error)
	GetRefreshToken(ctx context.Context, rawToken string) (models.RefreshToken, error)
	SetEmailVerified(ctx context.Context, uid int64) error
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL, refreshTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		log:         log,
		tokenTTL:    tokenTTL,
		refreshTTL:  refreshTTL,
	}
}

// * Login проверяет учетные данные и возвращает JWT и refresh token
func (a *Auth) Login(
	ctx context.Context,
	email, password string,
	appID int32,
) (accessToken string, refreshToken string, err error) {
	const op = "Auth.Login"
	log := a.log.With(slog.String("op", op))

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found")
			return "", "", ErrInvalidCredentials
		}

		log.Error("failed to get user", sl.Err(err))
		return "", "", err
	}

	if !user.IsVerified {
		return "", "", ErrEmailNotVerified
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", sl.Err(err))
		return "", "", ErrInvalidCredentials
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", "", ErrInvalidAppID
	}

	accessToken, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))
		return "", "", err
	}

	refreshTokenValue, err := jwt.NewRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", err
	}

	refreshHash, err := bcrypt.GenerateFromPassword([]byte(refreshTokenValue), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash refresh token", sl.Err(err))
		return "", "", err
	}

	err = a.usrSaver.SaveRefreshToken(ctx, user.ID, appID, refreshHash, time.Now().Add(a.refreshTTL))
	if err != nil {
		log.Error("failed to save refresh token", sl.Err(err))
		return "", "", err
	}

	log.Info("user logged in successfully", slog.Int64("uid", user.ID))
	return accessToken, refreshTokenValue, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	username string,
	pass string,
) (int64, error) {
	const op = "auth.registerNewUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("Registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, username, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("User already exists")

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("Failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Refresh(
	ctx context.Context,
	refreshToken string,
) (string, string, error) {
	const op = "auth.refresh"

	log := a.log.With(
		slog.String("op", op),
	)

	rt, err := a.usrProvider.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Warn("refresh token not found", sl.Err(err))
		return "", "", ErrInvalidCredentials
	}

	if time.Now().After(rt.ExpiresAt) {
		log.Warn("refresh token expired")

		return "", "", ErrInvalidCredentials
	}

	user, err := a.usrProvider.UserByID(ctx, rt.UserID)
	if err != nil {
		log.Error("failed to load user", sl.Err(err))
		return "", "", ErrInvalidCredentials
	}

	app, err := a.appProvider.App(ctx, rt.AppID)
	if err != nil {
		return "", "", ErrInvalidAppID
	}

	accessToken, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate access token", sl.Err(err))
		return "", "", err
	}

	newRefresh, err := jwt.NewRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", sl.Err(err))
		return "", "", err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newRefresh), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash new refresh token", sl.Err(err))
		return "", "", err
	}

	err = a.usrSaver.UpdateRefreshToken(
		ctx,
		rt.UserID,
		rt.TokenHash,
		newHash,
		time.Now().Add(a.refreshTTL),
	)
	if err != nil {
		log.Error("failed to update refresh token", sl.Err(err))
		return "", "", err
	}

	log.Info("refresh successful", slog.Int64("uid", user.ID))

	return accessToken, newRefresh, nil
}

func (a *Auth) VerifyUser(
	ctx context.Context,
	verificationToken string,
	verificationTokenSecret string,
) error {
	const op = "auth.VerifyUser"

	log := a.log.With(
		slog.String("op", op),
	)

	user_id, err := verification.ParseVerificationToken(verificationToken, verificationTokenSecret)
	if err != nil {
		log.Error("failed to update parse verification token", sl.Err(err))

		return err
	}

	if err = a.usrProvider.SetEmailVerified(ctx, user_id); err != nil {
		log.Error("failed to update update status in database", sl.Err(err))

		return err
	}

	return nil
}

func (a *Auth) Logout(
	ctx context.Context,
	rawRefreshToken string,
) error {
	const op = "auth.Logout"

	log := a.log.With(
		slog.String("op", op),
	)

	rt, err := a.usrProvider.GetRefreshToken(ctx, rawRefreshToken)
	if err != nil {
		log.Warn("refresh token not found", slog.Any("err", err))
		return ErrInvalidCredentials
	}

	err = a.usrSaver.DeleteRefreshToken(ctx, rt.TokenHash)
	if err != nil {
		log.Error("failed to delete refresh token", slog.Any("err", err))
		return err
	}

	log.Info("logout successful")

	return nil
}
