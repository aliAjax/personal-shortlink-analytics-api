package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/shortlink-api/pkg/httputil"
	"github.com/example/shortlink-api/pkg/jwtutil"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httputil.WriteError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			claims, err := jwtutil.Parse(secret, strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, usernameKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(r *http.Request) int64 {
	value, _ := r.Context().Value(userIDKey).(int64)
	return value
}

func Username(r *http.Request) string {
	value, _ := r.Context().Value(usernameKey).(string)
	return value
}
