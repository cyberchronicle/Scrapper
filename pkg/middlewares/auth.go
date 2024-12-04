package middlewares

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

const UserId = "User-Id"

func Auth(next http.Handler, checkAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx    context.Context
			userId int
			err    error
		)
		if checkAuth {
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
			userId, err = strconv.Atoi(userIdStr)
			if err != nil {
				http.Error(w, "User-Id header must be number", http.StatusBadRequest)
				return
			}
		}
		ctx = context.WithValue(r.Context(), UserId, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserId(ctx context.Context) (int, error) {
	userId, ok := ctx.Value(UserId).(int)
	if !ok {
		return 0, fmt.Errorf("User-Id is requried")
	}
	return userId, nil
}
