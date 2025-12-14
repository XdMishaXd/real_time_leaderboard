package verify

import (
	"auth_service/internal/auth"
	resp "auth_service/internal/lib/api/response"
	sl "auth_service/internal/lib/logger"
	"auth_service/internal/lib/verification"
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

func New(ctx context.Context,
	log *slog.Logger,
	authMiddleware auth.Auth,
	tokenSecret string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.verify.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		token := r.URL.Query().Get("token")
		if token == "" {
			log.Warn("missing verification token")

			render.JSON(w, r, resp.Error("missing token"))

			return
		}

		userID, err := verification.ParseVerificationToken(token, tokenSecret)
		if err != nil {
			log.Warn("invalid verification token", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid or expired token"))

			return
		}

		if err := authMiddleware.VerifyUser(ctx, token, tokenSecret); err != nil {
			log.Error("failed to mark user as verified", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("email verified successfully", slog.Int64("uid", userID))

		ResponseOK(w, r)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
