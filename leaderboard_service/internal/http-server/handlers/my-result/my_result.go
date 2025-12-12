package myresult

import (
	"context"
	resp "leaderboard_service/internal/lib/api/response"
	"leaderboard_service/internal/lib/jwt"
	sl "leaderboard_service/internal/lib/logger"
	"leaderboard_service/internal/models"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Game   string `json:"game"`
	UserID int64  `json:"user_id"`
	Score  int64  `json:"score"`
	Rank   int64  `json:"rank"`
}

type StatisticsGetter interface {
	GetUserScoreAndRank(ctx context.Context, game string, userID int64) (*models.UserRank, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	statsGetter StatisticsGetter,
	jwtParser jwt.JWTParser,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.my_result.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		game := chi.URLParam(r, "game")
		if game == "" {
			log.Info("Game is empty")

			render.JSON(w, r, resp.Error("not found"))

			return
		}

		userID, _, err := jwtParser.ParseToken(r.Header.Get("Authorization"))
		if err != nil {
			log.Error("Failed to parse JWT", sl.Err(err))

			render.JSON(w, r, resp.Error("Unauthorized"))

			return
		}

		userRank, err := statsGetter.GetUserScoreAndRank(ctx, game, userID)
		if err != nil {
			log.Error("Failed get user score", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("user statistics got successfully", slog.Int64("uid", userID))

		ResponseOK(w, r, game, userID, userRank.Score, userRank.Rank)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, game string, userID, score, rank int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		UserID:   userID,
		Game:     game,
		Score:    score,
		Rank:     rank,
	})
}
