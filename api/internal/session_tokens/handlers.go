package sessiontokens

import (
	"auth/internal/apperror"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
)

type IssueSessionTokensParams struct {
	UserID               ulid.ULID
	TokenSource          string
	ClientID             *ulid.ULID
	AuthorizationID      *ulid.ULID
	ParentRefreshTokenID *[]byte
}

type SessionTokens struct {
	AccessToken    string
	RefreshToken   string
	ExpiresIn      int
	RefreshTokenID ulid.ULID
}

func (s *SessionTokenService) IssueSessionTokens(ctx context.Context, params IssueSessionTokensParams) (SessionTokens, error) {
	if params.TokenSource != TokenSourceSelf && params.TokenSource != TokenSourceClient {
		return SessionTokens{}, apperror.NewBadRequest("Invalid token source")
	}
	if params.TokenSource == TokenSourceClient && (params.ClientID == nil || params.AuthorizationID == nil) {
		return SessionTokens{}, apperror.NewBadRequest("Client tokens require client and authorization")
	}
	if params.TokenSource == TokenSourceSelf && (params.ClientID != nil || params.AuthorizationID != nil) {
		return SessionTokens{}, apperror.NewBadRequest("Self tokens cannot include client authorization")
	}

	roles, err := s.roleRepo.GetByUserID(params.UserID)
	if err != nil {
		return SessionTokens{}, err
	}

	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey:  s.jwtSigningKey,
		keyID:       s.jwtSigningKeyID,
		issuer:      s.issuer,
		userID:      params.UserID,
		clientID:    params.ClientID,
		tokenSource: params.TokenSource,
		roles:       utils.Map(roles, func(role model.Roles) string { return role.Name }),
		expiry:      s.accessTokenExpiry,
	})
	if err != nil {
		return SessionTokens{}, apperror.NewInternalServerError("Token generation error")
	}
	events.Log(ctx, &s.eventRepo, events.AccessTokenCreated, &params.UserID, events.AccessTokenCreatedData{})

	clientInfo := clientInfoFromContext(ctx)
	refreshTokenID := ulid.Make()
	refreshTokenModel := model.RefreshTokens{
		ID:          refreshTokenID.Bytes(),
		UserID:      params.UserID.Bytes(),
		TokenSource: params.TokenSource,
		ParentID:    params.ParentRefreshTokenID,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(s.refreshTokenExpiry),
		RevokedAt:   nil,
		IPAddress:   clientInfo.IP,
		UserAgent:   clientInfo.UserAgent,
	}
	if params.ClientID != nil {
		clientID := params.ClientID.Bytes()
		refreshTokenModel.ClientID = &clientID
	}
	if params.AuthorizationID != nil {
		authorizationID := params.AuthorizationID.Bytes()
		refreshTokenModel.AuthorizationID = &authorizationID
	}
	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return SessionTokens{}, err
	}
	events.Log(ctx, &s.eventRepo, events.RefreshTokenCreated, &params.UserID, events.RefreshTokenCreatedData{
		RefreshTokenID: ulidutil.ToPrefixed("token", refreshTokenID),
	})

	refreshToken, err := GenerateRefreshToken(GenerateRefreshTokenParams{
		privateKey:  s.jwtSigningKey,
		keyID:       s.jwtSigningKeyID,
		issuer:      s.issuer,
		userID:      params.UserID,
		clientID:    params.ClientID,
		tokenSource: params.TokenSource,
		tokenID:     refreshTokenID,
		expiry:      s.refreshTokenExpiry,
	})
	if err != nil {
		return SessionTokens{}, apperror.NewInternalServerError("Token generation error")
	}

	return SessionTokens{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		ExpiresIn:      int(s.accessTokenExpiry.Seconds()),
		RefreshTokenID: refreshTokenID,
	}, nil
}

func clientInfoFromContext(ctx context.Context) httputil.ClientInfo {
	info, ok := httputil.ClientInfoFromContext(ctx)
	if !ok || info.IP == "" {
		return httputil.ClientInfo{IP: "0.0.0.0", UserAgent: "unknown"}
	}
	if info.UserAgent == "" {
		info.UserAgent = "unknown"
	}

	return info
}
