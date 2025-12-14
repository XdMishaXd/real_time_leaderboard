package postgres

import (
	"auth_service/internal/config"
	"auth_service/internal/models"
	"auth_service/internal/storage"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*PostgresRepo, error) {
	const op = "storage.postgres.New"

	dsn := dsn(cfg)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse config: %w", op, err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create pool: %w", op, err)
	}

	return &PostgresRepo{pool: pool}, nil
}

func (r *PostgresRepo) SaveUser(ctx context.Context, email, username string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	var id int64

	err := r.pool.QueryRow(ctx, query, email, username, string(passHash)).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgx.PgError); ok && pgErr.Code == "23505" {
			return 0, storage.ErrUserExists
		}

		return 0, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	return id, nil
}

func (r *PostgresRepo) User(ctx context.Context, email string) (models.User, error) {
	query := `
		SELECT id, email, username, password_hash, is_verified
		FROM users
		WHERE email = $1;
	`

	row := r.pool.QueryRow(ctx, query, email)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PassHash,
		&u.IsVerified,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}

	return u, err
}

func (r *PostgresRepo) UserByID(ctx context.Context, id int64) (models.User, error) {
	query := `
		SELECT id, email, username, password_hash, is_verified
		FROM users
		WHERE id = $1;
	`

	row := r.pool.QueryRow(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PassHash,
		&u.IsVerified,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}

	return u, err
}

func (r *PostgresRepo) SetEmailVerified(ctx context.Context, userID int64) error {
	query := `UPDATE users SET is_verified = TRUE WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, userID)

	return err
}

func (r *PostgresRepo) SaveRefreshToken(
	ctx context.Context,
	userID int64,
	appID int32,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, app_id ,token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, userID, appID, tokenHash, expiresAt)
	return err
}

func (r *PostgresRepo) UpdateRefreshToken(
	ctx context.Context,
	userID int64,
	oldTokenHash []byte,
	newTokenHash []byte,
	expiresAt time.Time,
) error {
	const query = `
		UPDATE refresh_tokens
		SET token_hash = $1, expires_at = $2
		WHERE user_id = $3 AND token_hash = $4
	`

	_, err := r.pool.Exec(ctx, query, newTokenHash, expiresAt, userID, oldTokenHash)
	return err
}

func (r *PostgresRepo) GetRefreshToken(ctx context.Context, rawToken string) (models.RefreshToken, error) {
	const query = `
		SELECT user_id, app_id, token_hash, expires_at
		FROM refresh_tokens
		WHERE expires_at > NOW();
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return models.RefreshToken{}, err
	}
	defer rows.Close()

	var rt models.RefreshToken

	for rows.Next() {
		var (
			userID    int64
			appID     int32
			tokenHash []byte
			expiresAt time.Time
		)

		err := rows.Scan(&userID, &appID, &tokenHash, &expiresAt)
		if err != nil {
			return models.RefreshToken{}, err
		}

		if bcrypt.CompareHashAndPassword(tokenHash, []byte(rawToken)) == nil {
			rt.UserID = userID
			rt.AppID = appID
			rt.TokenHash = tokenHash
			rt.ExpiresAt = expiresAt
			return rt, nil
		}
	}
	if rows.Err() != nil {
		return models.RefreshToken{}, rows.Err()
	}

	return models.RefreshToken{}, storage.ErrRefreshTokenNotFound
}

func (r *PostgresRepo) DeleteRefreshToken(ctx context.Context, tokenHash []byte) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, tokenHash)

	return err
}

func (r *PostgresRepo) App(ctx context.Context, appID int32) (models.App, error) {
	query := `
		SELECT id, name, secret
		FROM apps
		WHERE id = $1;
	`

	var a models.App

	err := r.pool.QueryRow(ctx, query, appID).Scan(&a.ID, &a.Name, &a.Secret)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.App{}, storage.ErrAppNotFound
	}

	return a, err
}

func (r *PostgresRepo) Close() {
	r.pool.Close()
}

// * dsn формирует конфигурацию базы данных.
func dsn(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s database=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)
}
