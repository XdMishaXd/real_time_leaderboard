package gametop

import (
	"context"
	resp "leaderboard_service/internal/lib/api/response"
	"leaderboard_service/internal/lib/jwt"
	sl "leaderboard_service/internal/lib/logger"
	"leaderboard_service/internal/models"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Game string                    `json:"game"`
	Top  []models.LeaderboardEntry `json:"top"`
}

type TopGetter interface {
	GetTop(ctx context.Context, game string, limit int64) ([]models.LeaderboardEntry, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	topGetter TopGetter,
	jwtParser jwt.JWTParser,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.sumbit.New"

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

		limitStr := r.URL.Query().Get("limit")
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit <= 0 {
			log.Error("limit is empty", sl.Err(err))

			render.JSON(w, r, resp.Error("Invalid limit"))

			return
		}

		_, _, err = jwtParser.ParseToken(r.Header.Get("Authorization"))
		if err != nil {
			log.Error("Failed to parse JWT", sl.Err(err))

			render.JSON(w, r, resp.Error("Unauthorized"))

			return
		}

		gameTop, err := topGetter.GetTop(ctx, game, limit)
		if err != nil {
			log.Error("Failed get top", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("game top got successfully", slog.String("game", game))

		ResponseOK(w, r, game, gameTop)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, game string, top []models.LeaderboardEntry) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Game:     game,
		Top:      top,
	})
}
