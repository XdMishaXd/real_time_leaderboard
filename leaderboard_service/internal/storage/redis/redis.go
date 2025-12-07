package redis

import (
	"context"
	"fmt"
	"leaderboard_service/internal/models"
	"leaderboard_service/internal/storage"

	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

const (
	gamesKey = "games"
)

func New(ctx context.Context, db int, addr string) (*RedisRepo, error) {
	const op = "storage.redis.New"

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RedisRepo{client: rdb}, nil
}

func (r *RedisRepo) EnsureGameExists(ctx context.Context, game string) error {
	_, err := r.client.Eval(ctx, storage.EnsureGameExistsScript, []string{game}).Result()
	if err != nil {
		if strings.Contains(err.Error(), "GAME_ALREADY_EXISTS") {
			return storage.ErrGameAlreadyExists
		}

		return err
	}

	return nil
}

func (r *RedisRepo) GetAllGames(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, gamesKey).Result()
}

func (r *RedisRepo) SubmitScore(ctx context.Context, game, userID string, score int64) error {
	_, err := r.client.Eval(ctx, storage.SubmitScoreScript, []string{game}, userID, score).Result()
	if err != nil {
		if strings.Contains(err.Error(), "GAME_NOT_FOUND") {
			return storage.ErrGameNotFound
		}

		return err
	}

	return err
}

func (r *RedisRepo) GetTop(ctx context.Context, game string, limit int64) ([]models.LeaderboardEntry, error) {
	key := leaderboardKey(game)

	// ZREVRANGE берет от наибольшего к наименьшему
	res, err := r.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]models.LeaderboardEntry, 0, len(res))

	for i, z := range res {
		entries = append(entries, models.LeaderboardEntry{
			UserID: fmt.Sprint(z.Member),
			Score:  int64(z.Score),
			Rank:   int64(i + 1),
		})
	}

	return entries, nil
}

func (r *RedisRepo) GetUserScoreAndRank(ctx context.Context, game, userID string) (*models.UserRank, error) {
	key := leaderboardKey(game)

	// ZSCORE для получения очков
	score, err := r.client.ZScore(ctx, key, userID).Result()
	if err == redis.Nil {
		return nil, nil // user not found
	}
	if err != nil {
		return nil, err
	}

	// ZREVRANK (от большего к меньшему)
	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &models.UserRank{
		Score: int64(score),
		Rank:  int64(rank + 1),
	}, nil
}

func (r *RedisRepo) Close() {
	r.client.Close()
}

func leaderboardKey(game string) string {
	return "leaderboard:" + game
}
