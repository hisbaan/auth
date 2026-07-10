package httputil

import (
	"net/http"
	"testing"
)

func TestClientInfoFromRequestIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "uses first forwarded IP",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "134.128.221.238, 44.223.87.191",
			},
			want: "134.128.221.238",
		},
		{
			name:       "skips invalid forwarded entries",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "unknown, 44.223.87.191",
			},
			want: "44.223.87.191",
		},
		{
			name:       "falls back to real IP",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "unknown",
				"X-Real-IP":       "44.223.87.191",
			},
			want: "44.223.87.191",
		},
		{
			name:       "uses remote addr host",
			remoteAddr: "10.0.0.1:1234",
			want:       "10.0.0.1",
		},
		{
			name:       "uses IPv6 forwarded IP",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1, 44.223.87.191",
			},
			want: "2001:db8::1",
		},
		{
			name:       "falls back when no valid IP exists",
			remoteAddr: "unknown",
			headers: map[string]string{
				"X-Forwarded-For": "unknown",
				"X-Real-IP":       "also-unknown",
			},
			want: "0.0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			got := ClientInfoFromRequest(req).IP
			if got != tt.want {
				t.Fatalf("ClientInfoFromRequest().IP = %q, want %q", got, tt.want)
			}
		})
	}
}
