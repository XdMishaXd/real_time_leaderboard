package submit

import (
	"context"
	resp "leaderboard_service/internal/lib/api/response"
	"leaderboard_service/internal/lib/jwt"
	sl "leaderboard_service/internal/lib/logger"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Game  string `json:"game" validate:"required"`
	Score int64  `json:"score" validate:"required"`
}

type Response struct {
	resp.Response
	Game  string `json:"game"`
	Score int64  `json:"score"`
	Rank  int64  `json:"rank"`
}

type ScoreSaver interface {
	SubmitScore(ctx context.Context, game, username string, userID, score int64) (int64, error)
}

func New(ctx context.Context, log *slog.Logger, scoreSaver ScoreSaver, jwtParser jwt.JWTParser) http.HandlerFunc {
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

		userID, username, err := jwtParser.ParseToken(r.Header.Get("Authorization"))
		if err != nil {
			log.Error("Failed to parse JWT", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		rank, err := scoreSaver.SubmitScore(ctx, req.Game, username, userID, req.Score)
		if err != nil {
			log.Error("Failed save score", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		ResponseOK(w, r, req.Game, req.Score, rank)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, game string, score, rank int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Game:     game,
		Score:    score,
		Rank:     rank,
	})
}
