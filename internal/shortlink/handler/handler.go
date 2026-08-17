package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/example/shortlink-api/internal/shortlink/model"
	"github.com/example/shortlink-api/internal/shortlink/service"
	"github.com/example/shortlink-api/pkg/httputil"
	"github.com/example/shortlink-api/pkg/middleware"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/links", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	links, err := h.service.List(context.Background(), middleware.UserID(r))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list links")
		return
	}
	responses := model.ToResponses(links)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"items": responses})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRequest
	if !httputil.DecodeJSON(w, r, &req) {
		return
	}
	link, err := h.service.Create(context.Background(), middleware.UserID(r), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateCode):
			httputil.WriteError(w, http.StatusConflict, "short code already exists")
		case errors.Is(err, service.ErrInvalidURL):
			httputil.WriteError(w, http.StatusBadRequest, "original_url must be a valid http or https URL")
		case errors.Is(err, service.ErrInvalidAlias):
			httputil.WriteError(w, http.StatusBadRequest, "custom_alias must be 3 to 32 letters, numbers, '-' or '_'")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "failed to create short link")
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, link.ToResponse(0))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	link, err := h.service.Get(context.Background(), id, middleware.UserID(r))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "short link not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get short link")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, link.ToResponse(0))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(context.Background(), id, middleware.UserID(r)); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "short link not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete short link")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
