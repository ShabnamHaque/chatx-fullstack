package middleware

import (
	"log"
	"net/http"
	"time"
)

func measureLatencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("API %s took %v", r.URL.Path, duration)
	})
}
