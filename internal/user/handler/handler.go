package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/shortlink-api/internal/user/model"
	"github.com/example/shortlink-api/internal/user/repository"
	"github.com/example/shortlink-api/internal/user/service"
	"github.com/example/shortlink-api/pkg/httputil"
	"github.com/example/shortlink-api/pkg/jwtutil"
)

type Handler struct {
	service      *service.Service
	jwtSecret    string
	expiresHours int
}

func NewHandler(service *service.Service, jwtSecret string, expiresHours int) *Handler {
	return &Handler{service: service, jwtSecret: jwtSecret, expiresHours: expiresHours}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if !httputil.DecodeJSON(w, r, &req) {
		return
	}
	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUsername) {
			httputil.WriteError(w, http.StatusConflict, "username already exists")
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	token, err := jwtutil.Generate(h.jwtSecret, user.ID, user.Username, h.expiresHours)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, model.AuthResponse{Token: token, User: user.ToResponse()})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if !httputil.DecodeJSON(w, r, &req) {
		return
	}
	user, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "login failed")
		return
	}
	token, err := jwtutil.Generate(h.jwtSecret, user.ID, user.Username, h.expiresHours)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, model.AuthResponse{Token: token, User: user.ToResponse()})
}
