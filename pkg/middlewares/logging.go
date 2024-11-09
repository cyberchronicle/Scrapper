package middlewares

import (
	"net/http"
	"time"

	log "github.com/rs/zerolog/log"
)

func Logger(name string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Info().Msgf("%s: %s %s %s", name, r.Method, r.RequestURI, time.Since(start))
		})
	}
}
