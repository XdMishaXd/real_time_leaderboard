package logout

import (
	"auth_service/internal/auth"
	resp "auth_service/internal/lib/api/response"
	sl "auth_service/internal/lib/logger"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Response struct {
	resp.Response
}

func New(
	ctx context.Context,
	log *slog.Logger,
	authMiddleware auth.Auth,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.logout.New"

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

		if err := authMiddleware.Logout(ctx, req.RefreshToken); err != nil {
			log.Error("failed to logout user", sl.Err(err))

			if errors.Is(auth.ErrInvalidCredentials, err) {
				render.JSON(w, r, resp.Error("invalid credentials"))

				return
			}

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("user logged out successfully")

		ResponseOK(w, r)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
