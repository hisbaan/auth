package wellknown

import (
	"auth/internal/utils/httputil"
	"net/http"
)

func (s *WellKnownService) GetJWKSHandler(w http.ResponseWriter, r *http.Request) {
	jwks := s.GetJWKS()
	httputil.JSONResponse(w, http.StatusOK, jwks)
}
