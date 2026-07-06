package auth

import (
	"auth/internal/apperror"
	"auth/internal/events"
	"auth/internal/jet/postgres/public/model"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/jwtutil"
	"auth/internal/utils/stringutil"
	"auth/internal/utils/tokenutil"
	"auth/internal/utils/ulidutil"
	"context"
	"crypto/ed25519"
	"net/http"
	"net/url"
	"time"

	"github.com/oklog/ulid/v2"
)

type CreateUserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	ReturnTo string `json:"return_to"`
}

func (s *AuthService) CreateUser(ctx context.Context, params CreateUserParams) error {
	username, err := stringutil.ValidateUsername(params.Username)
	if err != nil {
		return err
	}
	email, err := stringutil.ValidateUserEmail(params.Email, s.emailService.SenderAddress(), s.blockedEmailDomains)
	if err != nil {
		return err
	}
	if err := stringutil.ValidatePassword(params.Password); err != nil {
		return err
	}
	returnTo, err := s.validateEmailVerificationReturnTo(params.ReturnTo)
	if err != nil {
		return err
	}

	hash, err := HashPassword(params.Password)
	if err != nil {
		return err
	}

	user := model.Users{
		ID:            ulid.Make().Bytes(),
		Username:      username,
		Email:         email,
		EmailVerified: false,
		PasswordHash:  hash,
	}

	userExists, err := s.userRepo.WillConflict(user)
	if err != nil {
		return err
	}
	if userExists {
		return apperror.NewConflict("Username or email already in use")
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return err
	}

	userID := ulidutil.MustFromBytes(user.ID)
	events.Log(ctx, &s.eventRepo, events.UserCreated, &userID, events.UserCreatedData{})
	s.emailVerificationTokenRepo.RevokeByUserID(userID)

	token, hashedToken := tokenutil.Generate()
	emailVerificationTokenModel := model.EmailVerificationTokens{
		ID:        ulid.Make().Bytes(),
		UserID:    user.ID,
		TokenHash: hashedToken,
		Email:     email,
		ReturnTo:  returnTo,
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

	return nil
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTP     *int   `json:"totp,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Login(ctx context.Context, params LoginParams) (LoginResponse, error) {
	email, err := stringutil.NormalizeEmail(params.Email)
	if err != nil || params.Password == "" || len(params.Password) > stringutil.MaxPasswordLength {
		return LoginResponse{}, apperror.NewUnauthorized("Invalid credentials")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return LoginResponse{}, err
	}

	if user == nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationPasswordFailed, nil, events.AuthenticationPasswordFailedData{
			Email: email,
		})
		return LoginResponse{}, apperror.NewUnauthorized("Invalid credentials")
	}

	match := ComparePasswordAndHash(params.Password, user.PasswordHash)
	if !match {
		userID := ulidutil.MustFromBytes(user.ID)
		events.Log(ctx, &s.eventRepo, events.AuthenticationPasswordFailed, &userID, events.AuthenticationPasswordFailedData{
			Email: email,
		})
		return LoginResponse{}, apperror.NewUnauthorized("Invalid credentials")
	}
	if !user.EmailVerified {
		return LoginResponse{}, apperror.NewForbidden("Email verification required")
	}

	userID := ulidutil.MustFromBytes(user.ID)
	events.Log(ctx, &s.eventRepo, events.AuthenticationPasswordSucceeded, &userID, events.AuthenticationPasswordSucceededData{})
	tokens, err := s.sessionTokenService.IssueSessionTokens(ctx, sessiontokens.IssueSessionTokensParams{
		UserID:               userID,
		TokenSource:          sessiontokens.TokenSourceSelf,
		ParentRefreshTokenID: nil,
	})
	if err != nil {
		return LoginResponse{}, err
	}

	response := LoginResponse{
		AccessToken:  tokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		RefreshToken: tokens.RefreshToken,
	}

	return response, nil
}

func (s *AuthService) validateEmailVerificationReturnTo(value string) (*string, error) {
	if value == "" {
		return nil, nil
	}

	returnTo, err := url.Parse(value)
	if err != nil || !returnTo.IsAbs() {
		return nil, apperror.NewBadRequest("Invalid return_to")
	}
	issuer, err := url.Parse(s.issuer)
	if err != nil || !issuer.IsAbs() {
		return nil, apperror.NewInternalServerError("Invalid issuer")
	}
	if returnTo.Scheme != issuer.Scheme || returnTo.Host != issuer.Host || returnTo.Path != "/authorize" {
		return nil, apperror.NewBadRequest("Invalid return_to")
	}

	cleaned := returnTo.String()
	return &cleaned, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) {
	if token == "" {
		return
	}

	_, claims, err := jwtutil.ValidateToken(s.jwtSigningKey.Public().(ed25519.PublicKey), token, &sessiontokens.AccessClaims{})
	if err != nil {
		return
	}
	if err := jwtutil.ValidateClaims(claims.RegisteredClaims, s.issuer); err != nil {
		return
	}
	if claims.TokenType != "access" {
		return
	}

	userID, err := ulidutil.FromPrefixed("user", claims.Subject)
	if err != nil {
		return
	}
	events.Log(ctx, &s.eventRepo, events.AuthenticationLogout, &userID, events.AuthenticationLogoutData{})
}

type RefreshParams struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Refresh(ctx context.Context, params RefreshParams) (RefreshResponse, error) {
	tokens, err := s.sessionTokenService.RefreshSelfSession(ctx, params.RefreshToken)
	if err != nil {
		return RefreshResponse{}, err
	}

	return RefreshResponse{
		AccessToken:  tokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

type ForgotPasswordParams struct {
	Email string `json:"email"`
}

func (s *AuthService) ForgotPassword(ctx context.Context, params ForgotPasswordParams) error {
	email, err := stringutil.ValidateUserEmail(params.Email, s.emailService.SenderAddress(), s.blockedEmailDomains)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	userID := ulidutil.MustFromBytes(user.ID)

	s.passwordResetTokenRepo.RevokeByUserID(userID)

	token, hashedToken := tokenutil.Generate()
	passwordResetTokenModel := model.PasswordResetTokens{
		ID:        ulid.Make().Bytes(),
		UserID:    user.ID,
		TokenHash: hashedToken,
		ExpiresAt: time.Now().Add(time.Duration(15) * time.Minute),
		RevokedAt: nil,
		CreatedAt: time.Now(),
	}
	s.passwordResetTokenRepo.Create(passwordResetTokenModel)
	events.Log(ctx, &s.eventRepo, events.PasswordResetCreated, &userID, events.PasswordResetCreatedData{
		PasswordResetTokenID: ulidutil.ToPrefixed("password_reset_token", ulidutil.MustFromBytes(passwordResetTokenModel.ID)),
	})
	urlEncodedToken := tokenutil.URLEncode(token)

	go s.emailService.SendForgotPasswordEmail(email, user.Username, urlEncodedToken)

	return nil
}

type PasswordResetParams struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (s *AuthService) PasswordReset(ctx context.Context, params PasswordResetParams) error {
	if err := stringutil.ValidatePassword(params.NewPassword); err != nil {
		return err
	}

	token, err := tokenutil.URLDecode(params.Token)
	if err != nil {
		events.Log(ctx, &s.eventRepo, events.PasswordResetFailed, nil, events.PasswordResetFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return apperror.NewBadRequest("Invalid token")
	}

	hashedToken := tokenutil.Hash(token)
	passwordResetToken, err := s.passwordResetTokenRepo.GetByHash(hashedToken)
	if err != nil {
		if httpErr, ok := err.(apperror.HTTPError); ok && httpErr.StatusCode() == http.StatusNotFound {
			events.Log(ctx, &s.eventRepo, events.PasswordResetFailed, nil, events.PasswordResetFailedData{
				Reason: events.EventReasonInvalidToken,
			})
		}
		return err
	}

	userID := ulidutil.MustFromBytes(passwordResetToken.UserID)
	passwordResetTokenID := ulidutil.MustFromBytes(passwordResetToken.ID)
	passwordResetTokenIDValue := ulidutil.ToPrefixed("password_reset_token", passwordResetTokenID)
	if passwordResetToken.RevokedAt != nil {
		events.Log(ctx, &s.eventRepo, events.PasswordResetFailed, &userID, events.PasswordResetFailedData{
			PasswordResetTokenID: passwordResetTokenIDValue,
			Reason:               events.EventReasonRevokedToken,
		})
		return apperror.NewBadRequest("Invalid token")
	}
	if passwordResetToken.ExpiresAt.Before(time.Now()) {
		events.Log(ctx, &s.eventRepo, events.PasswordResetFailed, &userID, events.PasswordResetFailedData{
			PasswordResetTokenID: passwordResetTokenIDValue,
			Reason:               events.EventReasonExpiredToken,
		})
		return apperror.NewBadRequest("Invalid token")
	}

	if err := s.passwordResetTokenRepo.Revoke(passwordResetTokenID); err != nil {
		return err
	}

	hashedPassword, err := HashPassword(params.NewPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.SetPassword(userID, hashedPassword); err != nil {
		return err
	}
	if err := s.refreshTokenRepo.RevokeByUserID(userID); err != nil {
		return err
	}
	events.Log(ctx, &s.eventRepo, events.PasswordResetSucceeded, &userID, events.PasswordResetSucceededData{
		PasswordResetTokenID: passwordResetTokenIDValue,
	})

	return nil
}

type VerifyEmailParams struct {
	Token string `json:"token"`
}

type VerifyEmailResponse struct {
	ContinueURL *string `json:"continue_url,omitempty"`
}

func (s *AuthService) VerifyEmail(ctx context.Context, params VerifyEmailParams) (VerifyEmailResponse, error) {
	token, err := tokenutil.URLDecode(params.Token)
	if err != nil {
		events.Log(ctx, &s.eventRepo, events.AuthenticationVerifyEmailFailed, nil, events.AuthenticationVerifyEmailFailedData{
			Reason: events.EventReasonInvalidToken,
		})
		return VerifyEmailResponse{}, apperror.NewBadRequest("Invalid token")
	}

	hashedToken := tokenutil.Hash(token)

	verificationToken, err := s.emailVerificationTokenRepo.GetByHash(hashedToken)
	if err != nil {
		if httpErr, ok := err.(apperror.HTTPError); ok && httpErr.StatusCode() == http.StatusNotFound {
			events.Log(ctx, &s.eventRepo, events.AuthenticationVerifyEmailFailed, nil, events.AuthenticationVerifyEmailFailedData{
				Reason: events.EventReasonInvalidToken,
			})
		}
		return VerifyEmailResponse{}, err
	}

	if verificationToken.RevokedAt != nil {
		userID := ulidutil.MustFromBytes(verificationToken.UserID)
		events.Log(ctx, &s.eventRepo, events.AuthenticationVerifyEmailFailed, &userID, events.AuthenticationVerifyEmailFailedData{
			Reason: events.EventReasonRevokedToken,
		})
		return VerifyEmailResponse{}, apperror.NewBadRequest("Invalid token")
	}

	if verificationToken.ExpiresAt.Before(time.Now()) {
		userID := ulidutil.MustFromBytes(verificationToken.UserID)
		events.Log(ctx, &s.eventRepo, events.AuthenticationVerifyEmailFailed, &userID, events.AuthenticationVerifyEmailFailedData{
			Reason: events.EventReasonExpiredToken,
		})
		return VerifyEmailResponse{}, apperror.NewBadRequest("Invalid token")
	}

	tokenID := ulidutil.MustFromBytes(verificationToken.ID)
	if err := s.emailVerificationTokenRepo.Revoke(tokenID); err != nil {
		return VerifyEmailResponse{}, err
	}

	userID := ulidutil.MustFromBytes(verificationToken.UserID)
	if err := s.userRepo.UpdateEmail(userID, verificationToken.Email, true); err != nil {
		return VerifyEmailResponse{}, err
	}
	events.Log(ctx, &s.eventRepo, events.UserEmailVerified, &userID, events.UserEmailVerifiedData{})

	return VerifyEmailResponse{ContinueURL: verificationToken.ReturnTo}, nil
}
