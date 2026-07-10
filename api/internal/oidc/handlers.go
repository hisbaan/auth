package oidc

import (
	"auth/internal/apperror"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/sessions"
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/tokenutil"
	"auth/internal/utils/ulidutil"
	"context"
	"crypto/ed25519"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type AuthorizeParams struct {
	ResponseType string `json:"response_type"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	RawQuery     string `json:"-"`

	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`

	Scope string  `json:"scope"`
	State string  `json:"state"`
	Nonce *string `json:"nonce"`

	Prompt    string `json:"prompt"`
	LoginHint string `json:"login_hint"`
}

type AuthorizeResult struct {
	RedirectURI string
	Query       url.Values
}

type authorizePrompt string

const (
	authorizePromptDefault authorizePrompt = ""
	authorizePromptNone    authorizePrompt = "none"
	authorizePromptConsent authorizePrompt = "consent"
)

type authorizeRequest struct {
	client *model.Clients
	scopes []string
	state  string
	prompt authorizePrompt
}

func (s *OIDCService) Authorize(params AuthorizeParams, userID ulid.ULID) (AuthorizeResult, error) {
	authorizeRequest, result, err := s.validateAuthorizeRequest(params)
	if err != nil || authorizeRequest == nil {
		return result, err
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return AuthorizeResult{}, err
	}
	if !user.EmailVerified {
		result.Query.Set("error", "access_denied")
		result.Query.Set("error_description", "Email verification required")
		return result, nil
	}

	clientID := ulidutil.MustFromBytes(authorizeRequest.client.ID)
	authorization, err := s.userClientAuthorizationRepo.GetByUserIDAndClientID(userID, clientID)
	if err != nil {
		return AuthorizeResult{}, err
	}
	consentGranted := authorization != nil && authorization.RevokedAt == nil && hasGrantedScopes(authorization.GrantedScopes, authorizeRequest.scopes)
	if !consentGranted || authorizeRequest.prompt == authorizePromptConsent {
		if authorizeRequest.prompt == authorizePromptNone {
			result.Query.Set("error", "consent_required")
			result.Query.Set("error_description", "The request requires user consent")
			return result, nil
		}

		result.RedirectURI = strings.TrimRight(s.frontendURL, "/") + "/authorize"
		result.Query = url.Values{"request": []string{params.RawQuery}}
		return result, nil
	}

	code, codeHash := tokenutil.Generate()

	err = s.authorizationCodeRepo.Create(model.AuthorizationCodes{
		ID:                  ulid.Make().Bytes(),
		CodeHash:            codeHash,
		UserID:              userID.Bytes(),
		ClientID:            authorizeRequest.client.ID,
		RedirectURI:         authorizeRequest.client.RedirectURI,
		Scopes:              authorizeRequest.scopes,
		CodeChallenge:       params.CodeChallenge,
		CodeChallengeMethod: params.CodeChallengeMethod,
		Nonce:               params.Nonce,
		ExpiresAt:           time.Now().Add(time.Duration(15) * time.Minute),
		UsedAt:              nil,
	})
	if err != nil {
		return AuthorizeResult{}, err
	}

	result.Query.Set("code", tokenutil.URLEncode(code))
	return result, nil
}

type AuthorizeSession struct {
	UserID        ulid.ULID
	RotatedTokens *sessions.SessionTokens
}

func (s *OIDCService) ResolveAuthorizeSession(ctx context.Context, accessClaims *sessions.AccessClaims, refreshToken string) (*AuthorizeSession, error) {
	if accessClaims != nil {
		userID, err := ulidutil.FromPrefixed("user", accessClaims.Subject)
		if err != nil {
			return nil, apperror.NewUnauthorized("Invalid token")
		}
		return &AuthorizeSession{UserID: userID}, nil
	}

	if refreshToken == "" {
		return nil, nil
	}

	tokens, err := s.sessionTokenService.RefreshSelfSession(ctx, refreshToken)
	if err != nil {
		return nil, nil
	}

	return &AuthorizeSession{UserID: tokens.UserID, RotatedTokens: &tokens}, nil
}

func (s *OIDCService) validateAuthorizeRequest(params AuthorizeParams) (*authorizeRequest, AuthorizeResult, error) {
	clientID, err := ulidutil.FromPrefixed("client", params.ClientID)
	if err != nil {
		return nil, AuthorizeResult{}, err
	}

	client, err := s.clientRepo.GetByID(clientID)
	if err != nil {
		return nil, AuthorizeResult{}, err
	}
	if client == nil || client.RevokedAt != nil {
		return nil, AuthorizeResult{}, apperror.NewBadRequest("Invalid client")
	}
	if client.RedirectURI != params.RedirectURI {
		return nil, AuthorizeResult{}, apperror.NewBadRequest("Invalid redirect uri")
	}

	result := AuthorizeResult{RedirectURI: client.RedirectURI, Query: url.Values{}}
	result.Query.Set("iss", s.issuer)
	if params.State != "" {
		result.Query.Set("state", params.State)
	}

	scopes := strings.Fields(params.Scope)
	promptValue, err := parseAuthorizePrompt(params.Prompt)
	if err != nil {
		result.Query.Set("error", "invalid_request")
		result.Query.Set("error_description", err.Error())
		return nil, result, nil
	}

	if params.ResponseType != "code" {
		result.Query.Set("error", "unsupported_response_type")
		result.Query.Set("error_description", "Invalid response type")
		return nil, result, nil
	}
	if params.CodeChallengeMethod != "S256" {
		result.Query.Set("error", "invalid_request")
		result.Query.Set("error_description", "Invalid code challenge method")
		return nil, result, nil
	}
	if strings.TrimSpace(params.CodeChallenge) == "" {
		result.Query.Set("error", "invalid_request")
		result.Query.Set("error_description", "Invalid code challenge")
		return nil, result, nil
	}
	for _, scope := range scopes {
		if !slices.Contains(client.AllowedScopes, scope) {
			result.Query.Set("error", "invalid_scope")
			result.Query.Set("error_description", fmt.Sprintf("Invalid scope %s", scope))
			return nil, result, nil
		}
	}
	if !slices.Contains(scopes, "openid") {
		result.Query.Set("error", "invalid_scope")
		result.Query.Set("error_description", "Must contain openid scope")
		return nil, result, nil
	}

	return &authorizeRequest{client: client, scopes: scopes, state: params.State, prompt: promptValue}, result, nil
}

func parseAuthorizePrompt(prompt string) (authorizePrompt, error) {
	promptValues := strings.Fields(prompt)
	if len(promptValues) == 0 {
		return authorizePromptDefault, nil
	}
	if len(promptValues) > 1 {
		if slices.Contains(promptValues, string(authorizePromptNone)) {
			return authorizePromptDefault, fmt.Errorf("prompt=none cannot be combined with other prompt values")
		}
		return authorizePromptDefault, fmt.Errorf("only one prompt value is supported")
	}

	switch promptValues[0] {
	case string(authorizePromptNone):
		return authorizePromptNone, nil
	case string(authorizePromptConsent):
		return authorizePromptConsent, nil
	default:
		return authorizePromptDefault, fmt.Errorf("unsupported prompt %s", promptValues[0])
	}
}

func hasGrantedScopes(granted []string, requested []string) bool {
	for _, scope := range requested {
		if !slices.Contains(granted, scope) {
			return false
		}
	}
	return true
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

type TokenAuthorizationCodeParams struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

func (s *OIDCService) TokenAuthorizationCode(ctx context.Context, params TokenAuthorizationCodeParams) (TokenResponse, error) {
	if params.GrantType != "authorization_code" {
		return TokenResponse{}, NewUnsupportedGrantTypeTokenError("Invalid grant type")
	}
	if strings.TrimSpace(params.ClientID) == "" {
		return TokenResponse{}, NewInvalidRequestTokenError("Invalid client_id")
	}

	clientID, err := ulidutil.FromPrefixed("client", params.ClientID)
	if err != nil {
		return TokenResponse{}, NewInvalidRequestTokenError("Invalid client_id")
	}

	codeUrlDecode, err := tokenutil.URLDecode(params.Code)
	if err != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization code")
	}
	codeHash := tokenutil.Hash(codeUrlDecode)

	authorizationCode, err := s.authorizationCodeRepo.GetByCodeHash(codeHash)
	if err != nil {
		return TokenResponse{}, err
	}
	if authorizationCode == nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization code")
	}
	if authorizationCode.UsedAt != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization code")
	}
	if authorizationCode.ExpiresAt.Before(time.Now()) {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization code")
	}

	if authorizationCode.RedirectURI != params.RedirectURI {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid redirect URI")
	}
	if ulidutil.MustFromBytes(authorizationCode.ClientID) != clientID {
		return TokenResponse{}, NewInvalidClientTokenError("Invalid client")
	}

	if tokenutil.PKCES256Challenge(params.CodeVerifier) != authorizationCode.CodeChallenge {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid challenge verifier")
	}

	userID := ulidutil.MustFromBytes(authorizationCode.UserID)
	if _, err := s.requireActiveClient(clientID); err != nil {
		return TokenResponse{}, err
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return TokenResponse{}, err
	}
	if !user.EmailVerified {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization")
	}
	userClientAuthorization, err := s.userClientAuthorizationRepo.GetByUserIDAndClientID(userID, clientID)
	if err != nil {
		return TokenResponse{}, err
	}
	if userClientAuthorization == nil || userClientAuthorization.RevokedAt != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid authorization")
	}
	authorizationID := ulidutil.MustFromBytes(userClientAuthorization.ID)
	refreshToken, err := s.sessionTokenService.IssueRefreshToken(ctx, sessions.IssueRefreshTokenParams{
		UserID:               userID,
		TokenSource:          sessions.TokenSourceClient,
		ClientID:             &clientID,
		AuthorizationID:      &authorizationID,
		ParentRefreshTokenID: nil,
	})
	if err != nil {
		return TokenResponse{}, err
	}

	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey: s.jwtSigningKey,
		keyID:      s.jwtSigningKeyID,
		issuer:     s.issuer,
		userID:     userID,
		clientID:   clientID,
		scopes:     authorizationCode.Scopes,
		expiry:     s.accessTokenExpiry,
	})
	if err != nil {
		return TokenResponse{}, err
	}
	events.Log(ctx, &s.eventRepo, events.AccessTokenCreated, &userID, events.AccessTokenCreatedData{})

	idToken, err := GenerateIDToken(GenerateIDTokenParams{
		privateKey: s.jwtSigningKey,
		keyID:      s.jwtSigningKeyID,
		issuer:     s.issuer,
		userID:     userID,
		clientID:   clientID,
		user:       user,
		scopes:     authorizationCode.Scopes,
		nonce:      authorizationCode.Nonce,
		expiry:     s.idTokenExpiry,
	})
	if err != nil {
		return TokenResponse{}, err
	}
	if err := s.authorizationCodeRepo.MarkUsed(ulidutil.MustFromBytes(authorizationCode.ID)); err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTokenExpiry.Seconds()),
		RefreshToken: refreshToken.RefreshToken,
		Scope:        strings.Join(authorizationCode.Scopes, " "),
		IDToken:      idToken,
	}, nil
}

type TokenRefreshTokenParams struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

func (s *OIDCService) TokenRefreshToken(ctx context.Context, params TokenRefreshTokenParams) (TokenResponse, error) {
	if params.GrantType != "refresh_token" {
		return TokenResponse{}, NewUnsupportedGrantTypeTokenError("Invalid grant type")
	}
	if strings.TrimSpace(params.ClientID) == "" {
		return TokenResponse{}, NewInvalidRequestTokenError("Invalid client_id")
	}

	clientID, err := ulidutil.FromPrefixed("client", params.ClientID)
	if err != nil {
		return TokenResponse{}, NewInvalidRequestTokenError("Invalid client_id")
	}

	claims, err := jwtutil.ValidateToken(s.jwtSigningKey.Public().(ed25519.PublicKey), s.issuer, jwtutil.RefreshTokenJWTType, params.RefreshToken, &sessions.RefreshClaims{})
	if err != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}
	if claims.TokenSource != sessions.TokenSourceClient {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}
	if claims.ClientID == nil || *claims.ClientID != params.ClientID {
		return TokenResponse{}, NewInvalidClientTokenError("Invalid client")
	}

	tokenID, err := ulidutil.FromPrefixed("token", claims.ID)
	if err != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}

	refreshToken, err := s.refreshTokenRepo.GetByID(tokenID)
	if err != nil {
		return TokenResponse{}, err
	}
	if refreshToken == nil || refreshToken.RevokedAt != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}
	if refreshToken.TokenSource != sessions.TokenSourceClient || refreshToken.ClientID == nil || refreshToken.AuthorizationID == nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}
	if ulidutil.MustFromBytes(*refreshToken.ClientID) != clientID {
		return TokenResponse{}, NewInvalidClientTokenError("Invalid client")
	}
	if _, err := s.requireActiveClient(clientID); err != nil {
		return TokenResponse{}, err
	}

	authorizationID := ulidutil.MustFromBytes(*refreshToken.AuthorizationID)
	userClientAuthorization, err := s.userClientAuthorizationRepo.GetByID(authorizationID)
	if err != nil {
		return TokenResponse{}, err
	}
	if userClientAuthorization == nil || userClientAuthorization.RevokedAt != nil {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}

	refreshTokenUserID := ulidutil.MustFromBytes(refreshToken.UserID)
	user, err := s.userRepo.GetByID(refreshTokenUserID)
	if err != nil {
		return TokenResponse{}, err
	}
	if !user.EmailVerified {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}

	refreshTokenIDValue := ulidutil.ToPrefixed("token", tokenID)
	revoked, err := s.refreshTokenRepo.Revoke(tokenID)
	if err != nil {
		return TokenResponse{}, err
	}
	if !revoked {
		return TokenResponse{}, NewInvalidGrantTokenError("Invalid token")
	}
	events.Log(ctx, &s.eventRepo, events.RefreshTokenRevoked, &refreshTokenUserID, events.RefreshTokenRevokedData{
		RefreshTokenID: refreshTokenIDValue,
	})

	newRefreshToken, err := s.sessionTokenService.IssueRefreshToken(ctx, sessions.IssueRefreshTokenParams{
		UserID:               refreshTokenUserID,
		TokenSource:          sessions.TokenSourceClient,
		ClientID:             ulidutil.MustPtrFromBytes(refreshToken.ClientID),
		AuthorizationID:      ulidutil.MustPtrFromBytes(refreshToken.AuthorizationID),
		ParentRefreshTokenID: &refreshToken.ID,
	})
	if err != nil {
		return TokenResponse{}, err
	}

	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey: s.jwtSigningKey,
		keyID:      s.jwtSigningKeyID,
		issuer:     s.issuer,
		userID:     refreshTokenUserID,
		clientID:   clientID,
		scopes:     userClientAuthorization.GrantedScopes,
		expiry:     s.accessTokenExpiry,
	})
	if err != nil {
		return TokenResponse{}, err
	}
	events.Log(ctx, &s.eventRepo, events.AccessTokenCreated, &refreshTokenUserID, events.AccessTokenCreatedData{})

	idToken, err := GenerateIDToken(GenerateIDTokenParams{
		privateKey: s.jwtSigningKey,
		keyID:      s.jwtSigningKeyID,
		issuer:     s.issuer,
		userID:     refreshTokenUserID,
		clientID:   clientID,
		user:       user,
		scopes:     userClientAuthorization.GrantedScopes,
		nonce:      nil,
		expiry:     s.idTokenExpiry,
	})
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTokenExpiry.Seconds()),
		RefreshToken: newRefreshToken.RefreshToken,
		Scope:        strings.Join(userClientAuthorization.GrantedScopes, " "),
		IDToken:      idToken,
	}, nil
}

type RevokeTokenParams struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint"`
}

type AuthorizeConsentParams struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

func (s *OIDCService) GrantConsent(userID ulid.ULID, params AuthorizeConsentParams) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if !user.EmailVerified {
		return apperror.NewForbidden("Email verification required")
	}

	clientID, err := ulidutil.FromPrefixed("client", params.ClientID)
	if err != nil {
		return err
	}
	client, err := s.clientRepo.GetByID(clientID)
	if err != nil {
		return err
	}
	if client == nil || client.RevokedAt != nil {
		return apperror.NewBadRequest("Invalid client")
	}

	scopes := strings.Fields(params.Scope)
	if !slices.Contains(scopes, "openid") {
		return apperror.NewBadRequest("Must contain openid scope")
	}
	for _, scope := range scopes {
		if !slices.Contains(client.AllowedScopes, scope) {
			return apperror.NewBadRequest(fmt.Sprintf("Invalid scope %s", scope))
		}
	}

	_, err = s.userClientAuthorizationRepo.UpsertActive(userID, clientID, scopes)
	return err
}

type AuthorizeClientInfoResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	RedirectURI   string   `json:"redirect_uri"`
	AllowedScopes []string `json:"allowed_scopes"`
}

func (s *OIDCService) GetAuthorizeClientInfo(clientID string) (AuthorizeClientInfoResponse, error) {
	parsedClientID, err := ulidutil.FromPrefixed("client", clientID)
	if err != nil {
		return AuthorizeClientInfoResponse{}, err
	}
	client, err := s.clientRepo.GetByID(parsedClientID)
	if err != nil {
		return AuthorizeClientInfoResponse{}, err
	}
	if client == nil || client.RevokedAt != nil {
		return AuthorizeClientInfoResponse{}, apperror.NewBadRequest("Invalid client")
	}

	return AuthorizeClientInfoResponse{
		ID:            clientID,
		Name:          client.Name,
		RedirectURI:   client.RedirectURI,
		AllowedScopes: []string(client.AllowedScopes),
	}, nil
}

func (s *OIDCService) DenyAuthorize(params AuthorizeParams) (AuthorizeResult, error) {
	_, result, err := s.validateAuthorizeRequest(params)
	if err != nil {
		return AuthorizeResult{}, err
	}
	result.Query.Set("error", "access_denied")
	result.Query.Set("error_description", "The user denied the request")
	return result, nil
}

func (s *OIDCService) RevokeToken(ctx context.Context, params RevokeTokenParams) error {
	if strings.TrimSpace(params.Token) == "" {
		return apperror.NewBadRequest("Token is required")
	}

	claims, err := jwtutil.ValidateToken(s.jwtSigningKey.Public().(ed25519.PublicKey), s.issuer, jwtutil.RefreshTokenJWTType, params.Token, &sessions.RefreshClaims{})
	if err != nil {
		return nil
	}
	if claims.TokenSource != sessions.TokenSourceClient {
		return nil
	}

	tokenID, err := ulidutil.FromPrefixed("token", claims.ID)
	if err != nil {
		return nil
	}

	refreshToken, err := s.refreshTokenRepo.GetByID(tokenID)
	if err != nil {
		return err
	}
	if refreshToken == nil || refreshToken.RevokedAt != nil {
		return nil
	}

	if _, err := s.refreshTokenRepo.Revoke(tokenID); err != nil {
		return err
	}

	refreshTokenUserID := ulidutil.MustFromBytes(refreshToken.UserID)
	events.Log(ctx, &s.eventRepo, events.RefreshTokenRevoked, &refreshTokenUserID, events.RefreshTokenRevokedData{
		RefreshTokenID: ulidutil.ToPrefixed("token", tokenID),
	})

	return nil
}

type UserInfoResponse struct {
	Subject           string  `json:"sub"`
	PreferredUsername *string `json:"preferred_username,omitempty"`
	Email             *string `json:"email,omitempty"`
	EmailVerified     *bool   `json:"email_verified,omitempty"`
}

type userInfoIdentity struct {
	userID   ulid.ULID
	clientID ulid.ULID
	scopes   []string
}

func (s *OIDCService) UserInfo(bearerToken string) (UserInfoResponse, error) {
	identity, err := s.resolveUserInfoToken(bearerToken)
	if err != nil {
		return UserInfoResponse{}, err
	}

	client, err := s.clientRepo.GetByID(identity.clientID)
	if err != nil {
		return UserInfoResponse{}, err
	}
	if client == nil || client.RevokedAt != nil {
		return UserInfoResponse{}, apperror.NewUnauthorized("Invalid token")
	}

	authorization, err := s.userClientAuthorizationRepo.GetByUserIDAndClientID(identity.userID, identity.clientID)
	if err != nil {
		return UserInfoResponse{}, err
	}
	if authorization == nil || authorization.RevokedAt != nil {
		return UserInfoResponse{}, apperror.NewUnauthorized("Invalid token")
	}

	user, err := s.userRepo.GetByID(identity.userID)
	if err != nil {
		return UserInfoResponse{}, err
	}
	if user == nil {
		return UserInfoResponse{}, apperror.NewUnauthorized("Invalid token")
	}

	// A scope is honored only if the token was issued with it and the
	// authorization still grants it.
	hasScope := func(scope string) bool {
		return slices.Contains(identity.scopes, scope) && slices.Contains(authorization.GrantedScopes, scope)
	}

	response := UserInfoResponse{
		Subject: ulidutil.ToPrefixed("user", identity.userID),
	}
	if hasScope("profile") {
		response.PreferredUsername = &user.Username
	}
	if hasScope("email") {
		response.Email = &user.Email
		response.EmailVerified = &user.EmailVerified
	}

	return response, nil
}

func (s *OIDCService) resolveUserInfoToken(bearerToken string) (userInfoIdentity, error) {
	publicKey := s.jwtSigningKey.Public().(ed25519.PublicKey)

	claims, err := ValidateAccessToken(publicKey, s.issuer, bearerToken)
	if err != nil {
		return userInfoIdentity{}, apperror.NewUnauthorized("Invalid token")
	}

	userID, err := ulidutil.FromPrefixed("user", claims.Subject)
	if err != nil {
		return userInfoIdentity{}, apperror.NewUnauthorized("Invalid token")
	}
	clientID, err := ulidutil.FromPrefixed("client", claims.ClientID)
	if err != nil {
		return userInfoIdentity{}, apperror.NewUnauthorized("Invalid token")
	}

	return userInfoIdentity{userID: userID, clientID: clientID, scopes: claims.Scopes()}, nil
}

func (s *OIDCService) requireActiveClient(clientID ulid.ULID) (*model.Clients, error) {
	client, err := s.clientRepo.GetByID(clientID)
	if err != nil {
		return nil, err
	}
	if client == nil || client.RevokedAt != nil {
		return nil, NewInvalidClientTokenError("Invalid client")
	}

	return client, nil
}
