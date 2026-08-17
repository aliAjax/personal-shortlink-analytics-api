package handler

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/shortlink-api/internal/redirect/service"
	"github.com/example/shortlink-api/pkg/httputil"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/r/{code}", h.redirect)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing short code")
		return
	}
	target, err := h.service.Resolve(
		r.Context(),
		code,
		r.Referer(),
		r.UserAgent(),
		clientIP(r),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, "short link not found")
		case errors.Is(err, service.ErrExpired):
			httputil.WriteError(w, http.StatusGone, "short link expired")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve short link")
		}
		return
	}
	http.Redirect(w, r, target.OriginalURL, http.StatusFound)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
