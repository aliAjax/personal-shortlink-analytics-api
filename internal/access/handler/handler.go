package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	accessservice "github.com/example/shortlink-api/internal/access/service"
	shortlinkservice "github.com/example/shortlink-api/internal/shortlink/service"
	"github.com/example/shortlink-api/pkg/httputil"
	"github.com/example/shortlink-api/pkg/middleware"
)

type Handler struct {
	service          *accessservice.Service
	shortlinkService *shortlinkservice.Service
}

func NewHandler(service *accessservice.Service, shortlinkService *shortlinkservice.Service) *Handler {
	return &Handler{service: service, shortlinkService: shortlinkService}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/dashboard", h.dashboard)
	r.Get("/stats/links/{id}", h.linkStats)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Dashboard(context.Background(), middleware.UserID(r))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) linkStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	link, err := h.shortlinkService.Get(context.Background(), id, middleware.UserID(r))
	if err != nil {
		if errors.Is(err, shortlinkservice.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "short link not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get short link")
		return
	}
	stats, err := h.service.LinkStats(context.Background(), link.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}
	stats.ShortCode = link.ShortCode
	httputil.WriteJSON(w, http.StatusOK, stats)
}
