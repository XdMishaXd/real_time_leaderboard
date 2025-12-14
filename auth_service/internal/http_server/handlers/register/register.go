package register

import (
	"auth_service/internal/auth"
	resp "auth_service/internal/lib/api/response"
	sl "auth_service/internal/lib/logger"
	"auth_service/internal/lib/verification"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Pass     string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	UserID int64 `json:"user_id"`
}

func New(ctx context.Context,
	log *slog.Logger,
	authMiddleware auth.Auth,
	msgSender verification.Publisher,
	verificationTokenTTL time.Duration,
	verificationTokenSecret string,
	address string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"

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

		userID, err := authMiddleware.RegisterNewUser(ctx, req.Email, req.Username, req.Pass)
		if err != nil {
			log.Error("failed to register user", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		log.Info("User registered", slog.Int64("id", userID))

		err = verification.VerifyUserEmail(
			ctx,
			log,
			msgSender,
			verificationTokenTTL,
			verificationTokenSecret,
			userID,
			address,
			req.Email,
		)
		if err != nil {
			log.Error("Failed to send verification email", sl.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))

			return
		}

		ResponseOK(w, r, userID)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, userID int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		UserID:   userID,
	})
}
