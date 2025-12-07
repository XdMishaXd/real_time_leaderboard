package submit

import (
	"context"
	resp "leaderboard_service/internal/lib/api/response"
	"leaderboard_service/internal/lib/jwt"
	sl "leaderboard_service/internal/lib/logger"
	"leaderboard_service/internal/storage/redis"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Game  string `json:"game" validate:"required"`
	Score int    `json:"score" validate:"required"`
}

type Response struct {
	resp.Response
	Rank         int  `json:"rank"`
	Created_game bool `json:"created_game"`
}

func New(ctx context.Context, log *slog.Logger, scoreSaver redis.RedisRepo, jwtParser jwt.JWTParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.sumbit.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("Failed to decode request"))

			return
		}

		log.Info("Request body decoded")

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("Invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// if err := scoreSaver.SubmitScore(ctx, req.Game)

	}
}
