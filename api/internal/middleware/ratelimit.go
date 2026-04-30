package middleware

import (
	"auth/internal/utils/httputil"
	"net/http"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	entries := make(map[string]rateLimitEntry)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			key := httputil.ClientInfoFromRequest(r).IP
			if key == "" {
				key = r.RemoteAddr
			}

			mu.Lock()
			entry := entries[key]
			if entry.resetAt.IsZero() || now.After(entry.resetAt) {
				entry = rateLimitEntry{resetAt: now.Add(window)}
			}
			entry.count++
			entries[key] = entry
			limited := entry.count > limit
			mu.Unlock()

			if limited {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
