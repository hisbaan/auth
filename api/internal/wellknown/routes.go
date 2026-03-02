package wellknown

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *WellKnownService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		Get JWKS
	//	@Description	Returns the JSON Web Key Set (JWKS) containing public keys for verifying JWT tokens
	//	@Tags			wellknown
	//	@Success		200	{object}	JWKS
	//	@Router			/.well-known/jwks.json [get]
	r.Get("/jwks.json", s.GetJWKSHandler)

	return r
}
