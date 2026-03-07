package refreshtokens

import (
	"time"

	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils/ulidutil"

	"github.com/oklog/ulid/v2"
)

type ListRefreshTokensParams struct {
	Cursor string
	Limit  int
}

type RefreshTokenResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	Status    string     `json:"status"`
}

type ListRefreshTokensResponse struct {
	RefreshTokens []RefreshTokenResponse `json:"refresh_tokens"`
	NextCursor    string                 `json:"next_cursor"`
}

func (s *AdminRefreshTokensService) ListUserRefreshTokens(userID ulid.ULID, params ListRefreshTokensParams) (ListRefreshTokensResponse, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var cursor *ulid.ULID
	if params.Cursor != "" {
		c, err := ulidutil.FromPrefixed("token", params.Cursor)
		if err != nil {
			return ListRefreshTokensResponse{}, apperror.NewBadRequest("Invalid cursor")
		}
		cursor = &c
	}

	tokens, err := s.refreshTokenRepo.ListByUserID(userID, params.Limit+1, cursor)
	if err != nil {
		return ListRefreshTokensResponse{}, err
	}

	hasMore := len(tokens) > params.Limit
	if hasMore {
		tokens = tokens[:params.Limit]
	}

	result := make([]RefreshTokenResponse, len(tokens))
	for i, token := range tokens {
		result[i] = mapRefreshToken(token)
	}

	var nextCursor string
	if hasMore && len(tokens) > 0 {
		lastToken := tokens[len(tokens)-1]
		lastID, err := ulidutil.FromBytes(lastToken.ID)
		if err != nil {
			return ListRefreshTokensResponse{}, apperror.NewInternalServerError("Invalid refresh token ID")
		}
		nextCursor = ulidutil.ToPrefixed("token", lastID)
	}

	return ListRefreshTokensResponse{RefreshTokens: result, NextCursor: nextCursor}, nil
}

func mapRefreshToken(token model.RefreshTokens) RefreshTokenResponse {
	status := "active"
	if token.RevokedAt != nil {
		status = "revoked"
	} else if token.ExpiresAt.Before(time.Now()) {
		status = "expired"
	}

	return RefreshTokenResponse{
		ID:        ulidutil.ToPrefixed("token", ulidutil.MustFromBytes(token.ID)),
		UserID:    ulidutil.ToPrefixed("user", ulidutil.MustFromBytes(token.UserID)),
		IssuedAt:  token.IssuedAt,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		IPAddress: token.IPAddress,
		UserAgent: token.UserAgent,
		Status:    status,
	}
}
