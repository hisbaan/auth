package sessiontokens

import (
	"auth/internal/apperror"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils"
	"auth/internal/utils/httputil"
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/ulidutil"
	"context"
	"crypto/ed25519"
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
	UserID         ulid.ULID
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

	var roleNames []string
	if params.TokenSource == TokenSourceSelf {
		roles, err := s.roleRepo.GetByUserID(params.UserID)
		if err != nil {
			return SessionTokens{}, err
		}
		roleNames = utils.Map(roles, func(role model.Roles) string { return role.Name })
	}

	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey:  s.jwtSigningKey,
		keyID:       s.jwtSigningKeyID,
		issuer:      s.issuer,
		userID:      params.UserID,
		clientID:    params.ClientID,
		tokenSource: params.TokenSource,
		roles:       roleNames,
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
		UserID:         params.UserID,
	}, nil
}

func (s *SessionTokenService) RefreshSelfSession(ctx context.Context, refreshToken string) (SessionTokens, error) {
	publicKey := s.jwtSigningKey.Public().(ed25519.PublicKey)

	_, claims, err := jwtutil.ValidateToken(publicKey, refreshToken, &RefreshClaims{})
	if err != nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, nil, events.AuthenticationRefreshFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return SessionTokens{}, err
	}

	var userID *ulid.ULID
	if parsedUserID, err := ulidutil.FromPrefixed("user", claims.Subject); err == nil {
		userID = &parsedUserID
	}

	if claims.TokenType != "refresh" {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, userID, events.AuthenticationRefreshFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}
	if claims.TokenSource != TokenSourceSelf {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, userID, events.AuthenticationRefreshFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}
	if err := jwtutil.ValidateClaims(claims.RegisteredClaims, s.issuer); err != nil {
		reason := events.EventReasonInvalidToken
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			reason = events.EventReasonExpiredToken
		}
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, userID, events.AuthenticationRefreshFailedData{
			Reason: reason,
		})
		return SessionTokens{}, err
	}

	tokenID, err := ulidutil.FromPrefixed("token", claims.ID)
	if err != nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, userID, events.AuthenticationRefreshFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}

	refreshTokenModel, err := s.refreshTokenRepo.GetByID(tokenID)
	if err != nil {
		return SessionTokens{}, err
	}

	refreshTokenIDValue := ulidutil.ToPrefixed("token", tokenID)
	if refreshTokenModel == nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, userID, events.AuthenticationRefreshFailedData{
			RefreshTokenID: refreshTokenIDValue,
			Reason:         events.EventReasonUnknownRefreshToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}

	refreshTokenUserID := ulidutil.MustFromBytes(refreshTokenModel.UserID)
	if refreshTokenModel.RevokedAt != nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, &refreshTokenUserID, events.AuthenticationRefreshFailedData{
			RefreshTokenID: refreshTokenIDValue,
			Reason:         events.EventReasonRevokedToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}

	refreshTokenULID := ulidutil.MustFromBytes(refreshTokenModel.ID)
	revoked, err := s.refreshTokenRepo.Revoke(refreshTokenULID)
	if err != nil {
		return SessionTokens{}, err
	}
	if !revoked {
		events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshFailed, &refreshTokenUserID, events.AuthenticationRefreshFailedData{
			RefreshTokenID: refreshTokenIDValue,
			Reason:         events.EventReasonRevokedToken,
		})
		return SessionTokens{}, apperror.NewUnauthorized("Invalid token")
	}
	events.Log(ctx, &s.eventRepo, events.RefreshTokenRevoked, &refreshTokenUserID, events.RefreshTokenRevokedData{
		RefreshTokenID: refreshTokenIDValue,
	})

	tokens, err := s.IssueSessionTokens(ctx, IssueSessionTokensParams{
		UserID:               refreshTokenUserID,
		TokenSource:          TokenSourceSelf,
		ClientID:             nil,
		AuthorizationID:      nil,
		ParentRefreshTokenID: &refreshTokenModel.ID,
	})
	if err != nil {
		return SessionTokens{}, err
	}
	events.Log(ctx, &s.eventRepo, events.AuthenticationRefreshTokenRotated, &refreshTokenUserID, events.AuthenticationRefreshTokenRotatedData{
		OldRefreshTokenID: ulidutil.ToPrefixed("token", refreshTokenULID),
		NewRefreshTokenID: ulidutil.ToPrefixed("token", tokens.RefreshTokenID),
	})

	return tokens, nil
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
