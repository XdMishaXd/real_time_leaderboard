package login

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
	Email string `json:"email" validate:"required,email"`
	Pass  string `json:"password" validate:"required"`
	AppID int32  `json:"app_id" validate:"required"`
}

type Response struct {
	resp.Response
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(ctx context.Context,
	log *slog.Logger,
	authMiddleware auth.Auth,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.New"

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

		accessToken, refreshToken, err := authMiddleware.Login(ctx, req.Email, req.Pass, req.AppID)
		if err != nil {
			if errors.Is(auth.ErrInvalidCredentials, err) {
				render.JSON(w, r, resp.Error("Invalid credentials"))

				return
			}
			if errors.Is(auth.ErrInvalidAppID, err) {
				render.JSON(w, r, resp.Error("Invalid app id"))

				return
			}
			if errors.Is(auth.ErrEmailNotVerified, err) {
				render.JSON(w, r, resp.Error("email is not verified"))

				return
			}

			log.Error("failed to login user", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("User logged in successfully")

		ResponseOK(w, r, accessToken, refreshToken)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
