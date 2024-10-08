package middlewares

import (
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"
)

func Logger(name string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			zlog.Info().Msgf("%s: %s %s %s", name, r.Method, r.RequestURI, time.Since(start))
		})
	}
}
