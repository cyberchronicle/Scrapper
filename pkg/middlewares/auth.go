package middlewares

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

const UserId = "User-Id"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdStr := r.Header.Get("User-Id")
		if userIdStr == "" {
			http.Error(w, "User-Id header is required", http.StatusBadRequest)
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("remote_addr", r.RemoteAddr).
				Msgf("Unauthorized request: User-Id header is missing")
			return
		}
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			http.Error(w, "User-Id header must be number", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), UserId, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
