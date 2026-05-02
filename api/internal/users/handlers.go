package users

import (
	"auth/internal/apperror"
	"auth/internal/auth"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils"
	"auth/internal/utils/stringutil"
	"auth/internal/utils/tokenutil"
	"auth/internal/utils/ulidutil"
	"context"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

type GetUserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	Roles         []string  `json:"roles"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *UsersService) GetUser(userID ulid.ULID) (GetUserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	roles, err := s.roleRepo.GetByUserID(userID)
	if err != nil {
		return GetUserResponse{}, err
	}

	return GetUserResponse{
		ID:            ulidutil.ToPrefixed("user", userID),
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		Roles:         utils.Map(roles, func(role model.Roles) string { return role.Name }),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

type UpdateUserParams struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (s *UsersService) UpdateUser(ctx context.Context, userID ulid.ULID, params UpdateUserParams) error {
	username, err := stringutil.ValidateUsername(params.Username)
	if err != nil {
		return err
	}
	email, err := stringutil.NormalizeEmail(params.Email)
	if err != nil {
		return err
	}
	if email == s.emailService.SenderAddress() {
		return apperror.NewBadRequest("Invalid email")
	}

	user := model.Users{
		ID:       userID.Bytes(),
		Email:    email,
		Username: username,
	}

	willConflict, err := s.userRepo.WillConflict(user)
	if err != nil {
		return err
	}
	if willConflict {
		return apperror.NewConflict("Username or email already in use")
	}

	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if existing.Email != email {
		user.EmailVerified = false

		userID := ulidutil.MustFromBytes(user.ID)
		events.Log(ctx, &s.eventRepo, events.UserEmailVerificationRevoked, &userID, events.UserEmailVerificationRevokedData{
			Email: existing.Email,
		})
		s.emailVerificationTokenRepo.RevokeByUserID(userID)

		token, hashedToken := tokenutil.Generate()
		emailVerificationTokenModel := model.EmailVerificationTokens{
			ID:        ulid.Make().Bytes(),
			UserID:    user.ID,
			TokenHash: hashedToken,
			ExpiresAt: time.Now().Add(time.Duration(24) * time.Hour),
			RevokedAt: nil,
			CreatedAt: time.Now(),
		}
		s.emailVerificationTokenRepo.Create(emailVerificationTokenModel)
		events.Log(ctx, &s.eventRepo, events.UserEmailVerificationCreated, &userID, events.UserEmailVerificationCreatedData{
			Email: email,
		})
		urlEncodedToken := tokenutil.URLEncode(token)

		go s.emailService.SendVerifyEmail(email, username, urlEncodedToken)
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// TODO include update fields in event data
	events.Log(ctx, &s.eventRepo, events.UserUpdated, &userID, events.UserUpdatedData{})

	return nil
}

type UpdatePasswordParams struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *UsersService) UpdatePassword(ctx context.Context, userID ulid.ULID, params UpdatePasswordParams) error {
	if params.CurrentPassword == "" || len(params.CurrentPassword) > stringutil.MaxPasswordLength {
		return apperror.NewUnauthorized("Unauthorized")
	}
	if err := stringutil.ValidatePassword(params.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if !auth.ComparePasswordAndHash(params.CurrentPassword, user.PasswordHash) {
		return apperror.NewUnauthorized("Unauthorized")
	}

	passwordHash, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.SetPassword(userID, passwordHash)
	if err != nil {
		return err
	}

	err = s.refreshTokenRepo.RevokeByUserID(userID)
	if err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.UserPasswordChanged, &userID, events.UserPasswordChangedData{})

	return nil
}

func (s *UsersService) DeleteUser(ctx context.Context, userID ulid.ULID) error {
	events.Log(ctx, &s.eventRepo, events.UserDeleted, &userID, events.UserDeletedData{})

	return s.userRepo.Delete(userID)
}

type ClientResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          string     `json:"name"`
	RedirectURI   string     `json:"redirect_uri"`
	AllowedScopes []string   `json:"allowed_scopes"`
	RevokedAt     *time.Time `json:"revoked_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ListClientsParams struct {
	Cursor string
	Limit  int
}

type ListClientsResponse struct {
	Clients    []ClientResponse `json:"clients"`
	NextCursor string           `json:"next_cursor"`
}

func (s *UsersService) ListClients(userID ulid.ULID, params ListClientsParams) (ListClientsResponse, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var cursor *ulid.ULID
	if params.Cursor != "" {
		c, err := ulidutil.FromPrefixed("client", params.Cursor)
		if err != nil {
			return ListClientsResponse{}, apperror.NewBadRequest("Invalid cursor")
		}
		cursor = &c
	}

	clients, err := s.clientRepo.ListByUserID(userID, params.Limit+1, cursor)
	if err != nil {
		return ListClientsResponse{}, err
	}

	hasMore := len(clients) > params.Limit
	if hasMore {
		clients = clients[:params.Limit]
	}

	result := make([]ClientResponse, len(clients))
	for i, client := range clients {
		result[i] = ClientResponse{
			ID:            ulidutil.ToPrefixed("client", ulidutil.MustFromBytes(client.ID)),
			UserID:        ulidutil.ToPrefixed("user", ulidutil.MustFromBytes(client.UserID)),
			Name:          client.Name,
			RedirectURI:   client.RedirectURI,
			AllowedScopes: []string(client.AllowedScopes),
			RevokedAt:     client.RevokedAt,
			CreatedAt:     client.CreatedAt,
			UpdatedAt:     client.UpdatedAt,
		}
	}

	var nextCursor string
	if hasMore && len(clients) > 0 {
		lastClient := clients[len(clients)-1]
		nextCursor = ulidutil.ToPrefixed("client", ulidutil.MustFromBytes(lastClient.ID))
	}

	return ListClientsResponse{Clients: result, NextCursor: nextCursor}, nil
}

type ClientAuthorizationResponse struct {
	ClientID         string    `json:"client_id"`
	Name             string    `json:"name"`
	RedirectURI      string    `json:"redirect_uri"`
	GrantedScopes    []string  `json:"granted_scopes"`
	LastAuthorizedAt time.Time `json:"last_authorized_at"`
}

type ListClientAuthorizationsResponse struct {
	Authorizations []ClientAuthorizationResponse `json:"authorizations"`
}

func (s *UsersService) ListClientAuthorizations(userID ulid.ULID) (ListClientAuthorizationsResponse, error) {
	authorizations, err := s.userClientAuthorizationRepo.ListActiveWithClientByUserID(userID)
	if err != nil {
		return ListClientAuthorizationsResponse{}, err
	}

	response := make([]ClientAuthorizationResponse, 0, len(authorizations))
	for _, authorization := range authorizations {
		response = append(response, ClientAuthorizationResponse{
			ClientID:         ulidutil.ToPrefixed("client", ulidutil.MustFromBytes(authorization.ClientID)),
			Name:             authorization.ClientName,
			RedirectURI:      authorization.ClientRedirectURI,
			GrantedScopes:    []string(authorization.GrantedScopes),
			LastAuthorizedAt: authorization.LastAuthorizedAt,
		})
	}

	return ListClientAuthorizationsResponse{Authorizations: response}, nil
}

type ClientParams struct {
	Name          string   `json:"name"`
	RedirectURI   string   `json:"redirect_uri"`
	AllowedScopes []string `json:"allowed_scopes"`
}

func (s *UsersService) CreateClient(ctx context.Context, userID ulid.ULID, params ClientParams) error {
	normalized, err := validateClientParams(params.Name, params.RedirectURI, params.AllowedScopes)
	if err != nil {
		return err
	}

	clientID := ulid.Make()
	client := model.Clients{
		ID:            clientID.Bytes(),
		UserID:        userID.Bytes(),
		Name:          normalized.Name,
		RedirectURI:   normalized.RedirectURI,
		AllowedScopes: normalized.AllowedScopes,
	}

	if err := s.clientRepo.Create(client); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.ClientCreated, &userID, events.ClientCreatedData{
		ClientID: ulidutil.ToPrefixed("client", clientID),
		Name:     client.Name,
	})

	_, err = s.clientRepo.GetByIDAndUserID(clientID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersService) UpdateClient(ctx context.Context, userID ulid.ULID, clientID ulid.ULID, params ClientParams) error {
	client, err := s.clientRepo.GetByIDAndUserID(clientID, userID)
	if err != nil {
		return err
	}
	if client == nil {
		return apperror.NewNotFound("Client not found")
	}
	if client.RevokedAt != nil {
		return apperror.NewConflict("Client is revoked")
	}

	normalized, err := validateClientParams(params.Name, params.RedirectURI, params.AllowedScopes)
	if err != nil {
		return err
	}

	securityPropertiesChanged := client.RedirectURI != normalized.RedirectURI || !slices.Equal([]string(client.AllowedScopes), []string(normalized.AllowedScopes))

	client.Name = normalized.Name
	client.RedirectURI = normalized.RedirectURI
	client.AllowedScopes = normalized.AllowedScopes

	if err := s.clientRepo.Update(*client); err != nil {
		return err
	}
	if securityPropertiesChanged {
		if err := s.userClientAuthorizationRepo.RevokeByClientID(clientID); err != nil {
			return err
		}
		if err := s.refreshTokenRepo.RevokeByClientID(clientID); err != nil {
			return err
		}
	}

	events.Log(ctx, &s.eventRepo, events.ClientUpdated, &userID, events.ClientUpdatedData{
		ClientID:      ulidutil.ToPrefixed("client", clientID),
		Name:          client.Name,
		RedirectURI:   client.RedirectURI,
		AllowedScopes: client.AllowedScopes,
	})

	return nil
}

func (s *UsersService) RevokeClient(ctx context.Context, userID ulid.ULID, clientID ulid.ULID) error {
	client, err := s.clientRepo.GetByIDAndUserID(clientID, userID)
	if err != nil {
		return err
	}
	if client == nil {
		return apperror.NewNotFound("Client not found")
	}
	if client.RevokedAt != nil {
		return nil
	}

	if err := s.clientRepo.Revoke(clientID); err != nil {
		return err
	}
	if err := s.userClientAuthorizationRepo.RevokeByClientID(clientID); err != nil {
		return err
	}
	if err := s.refreshTokenRepo.RevokeByClientID(clientID); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.ClientRevoked, &userID, events.ClientRevokedData{
		ClientID: ulidutil.ToPrefixed("client", clientID),
	})

	return nil
}

func (s *UsersService) RevokeClientAuthorization(ctx context.Context, userID ulid.ULID, clientID ulid.ULID) error {
	authorization, err := s.userClientAuthorizationRepo.GetByUserIDAndClientID(userID, clientID)
	if err != nil {
		return err
	}
	if authorization == nil || authorization.RevokedAt != nil {
		return nil
	}

	authorizationID := ulidutil.MustFromBytes(authorization.ID)
	if err := s.userClientAuthorizationRepo.Revoke(authorizationID); err != nil {
		return err
	}
	if err := s.refreshTokenRepo.RevokeByAuthorizationID(authorizationID); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.ClientAuthorizationRevoked, &userID, events.ClientAuthorizationRevokedData{
		ClientID: ulidutil.ToPrefixed("client", clientID),
	})

	return nil
}

func (s *UsersService) DeleteClient(ctx context.Context, userID ulid.ULID, clientID ulid.ULID) error {
	client, err := s.clientRepo.GetByIDAndUserID(clientID, userID)
	if err != nil {
		return err
	}
	if client == nil {
		return apperror.NewNotFound("Client not found")
	}

	if err := s.clientRepo.Delete(clientID); err != nil {
		return err
	}
	if err := s.userClientAuthorizationRepo.RevokeByClientID(clientID); err != nil {
		return err
	}
	if err := s.refreshTokenRepo.RevokeByClientID(clientID); err != nil {
		return err
	}

	events.Log(ctx, &s.eventRepo, events.ClientDeleted, &userID, events.ClientDeletedData{
		ClientID: ulidutil.ToPrefixed("client", clientID),
	})

	return nil
}

func validateClientParams(name string, redirectURI string, allowedScopes []string) (ClientParams, error) {
	name, err := stringutil.NonEmpty(name)
	if err != nil {
		return ClientParams{}, apperror.NewBadRequest("Name must not be empty")
	}

	redirectURI, err = stringutil.NonEmpty(redirectURI)
	if err != nil {
		return ClientParams{}, apperror.NewBadRequest("RedirectURI must not be empty")
	}

	parsed, err := url.ParseRequestURI(redirectURI)
	if err != nil {
		return ClientParams{}, apperror.NewBadRequest("Redirect URI must be a valid URL")
	}
	if !isAllowedPublicClientRedirectURI(parsed) {
		return ClientParams{}, apperror.NewBadRequest("Redirect URI must use https, loopback http, or a custom app scheme")
	}

	normalizedScopes := make([]string, 0, len(allowedScopes)+1)
	seen := make(map[string]struct{}, len(allowedScopes)+1)
	for _, scope := range allowedScopes {
		scope, err = stringutil.NonEmpty(scope)
		if err != nil {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalizedScopes = append(normalizedScopes, scope)
	}

	if _, ok := seen["openid"]; !ok {
		normalizedScopes = append([]string{"openid"}, normalizedScopes...)
	}

	return ClientParams{
		Name:          name,
		RedirectURI:   parsed.String(),
		AllowedScopes: pq.StringArray(normalizedScopes),
	}, nil
}

func isAllowedPublicClientRedirectURI(parsed *url.URL) bool {
	if parsed == nil || parsed.Scheme == "" {
		return false
	}

	switch parsed.Scheme {
	case "https":
		return parsed.Host != ""
	case "http":
		host := parsed.Hostname()
		if host == "" {
			return false
		}
		return isLoopbackHost(host)
	default:
		return parsed.Host != "" || parsed.Opaque != "" || parsed.Path != ""
	}
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
