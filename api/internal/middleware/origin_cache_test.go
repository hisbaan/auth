package middleware

import (
	"errors"
	"testing"
	"time"
)

func TestOriginFromURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "strips path and query",
			raw:  "https://app.example.com/callback?foo=bar",
			want: "https://app.example.com",
		},
		{
			name: "keeps non-default port",
			raw:  "http://localhost:5173/cb",
			want: "http://localhost:5173",
		},
		{
			name: "strips default https port",
			raw:  "https://app.example.com:443/cb",
			want: "https://app.example.com",
		},
		{
			name: "strips default http port",
			raw:  "http://app.example.com:80/cb",
			want: "http://app.example.com",
		},
		{
			name: "lowercases scheme and host",
			raw:  "HTTPS://App.Example.COM/cb",
			want: "https://app.example.com",
		},
		{
			name: "rejects missing scheme",
			raw:  "app.example.com/cb",
			want: "",
		},
		{
			name: "rejects empty",
			raw:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := originFromURL(tt.raw); got != tt.want {
				t.Errorf("originFromURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestOriginCacheIsAllowed(t *testing.T) {
	calls := 0
	cache := NewOriginCache(func() ([]string, error) {
		calls++
		return []string{
			"https://app.example.com/callback",
			"http://localhost:5173/cb",
		}, nil
	}, time.Minute)

	if !cache.IsAllowed("https://app.example.com") {
		t.Error("expected registered client origin to be allowed")
	}
	if !cache.IsAllowed("http://localhost:5173") {
		t.Error("expected registered localhost origin to be allowed")
	}
	if cache.IsAllowed("https://evil.example.com") {
		t.Error("expected unregistered origin to be rejected")
	}
	if cache.IsAllowed("") {
		t.Error("expected empty origin to be rejected")
	}

	// The fetch result should be cached within the TTL rather than re-queried
	// on every call.
	if calls != 1 {
		t.Errorf("expected fetch to run once within TTL, ran %d times", calls)
	}
}

func TestOriginCacheInvalidateForcesRefetch(t *testing.T) {
	uris := []string{"https://app.example.com/cb"}
	calls := 0
	cache := NewOriginCache(func() ([]string, error) {
		calls++
		return uris, nil
	}, time.Hour) // long TTL so only Invalidate can trigger a refetch

	if !cache.IsAllowed("https://app.example.com") {
		t.Fatal("expected initial origin to be allowed")
	}
	if cache.IsAllowed("https://new.example.com") {
		t.Fatal("expected new origin to be rejected before it is registered")
	}
	if calls != 1 {
		t.Fatalf("expected one fetch within TTL, got %d", calls)
	}

	// A new client registers; without invalidation the long TTL would hide it.
	uris = append(uris, "https://new.example.com/cb")
	cache.Invalidate()

	if !cache.IsAllowed("https://new.example.com") {
		t.Error("expected newly registered origin to be allowed after Invalidate")
	}
	if calls != 2 {
		t.Errorf("expected Invalidate to force a refetch, got %d fetches", calls)
	}
}

func TestOriginCacheServesStaleOnError(t *testing.T) {
	fail := false
	cache := NewOriginCache(func() ([]string, error) {
		if fail {
			return nil, errors.New("db down")
		}
		return []string{"https://app.example.com/cb"}, nil
	}, -time.Second) // negative TTL forces a refresh on every call

	if !cache.IsAllowed("https://app.example.com") {
		t.Fatal("expected origin allowed after initial load")
	}

	// Subsequent fetches fail; the previously cached set must still be served.
	fail = true
	if !cache.IsAllowed("https://app.example.com") {
		t.Error("expected stale cache to be served when refresh fails")
	}
}
