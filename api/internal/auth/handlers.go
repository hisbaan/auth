package auth

import (
	"auth/internal/apperror"
	"auth/internal/jet/postgres/public/model"
	"auth/internal/ulidutil"
	"crypto/ed25519"
	"time"

	"github.com/oklog/ulid/v2"
)

type CreateUserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) CreateUser(params CreateUserParams) error {
	hash, err := HashPassword(params.Password)
	if err != nil {
		return err
	}

	user := model.Users{
		ID:           ulid.Make().Bytes(),
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hash,
	}

	return s.userRepo.Create(user)
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

func (s *AuthService) Login(params LoginParams, ip string, userAgent string) (LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(params.Email)
	if err != nil {
		return LoginResponse{}, err
	}

	if user == nil {
		return LoginResponse{}, apperror.NewUnauthorized("Invalid credentials")
	}

	match := ComparePasswordAndHash(params.Password, user.PasswordHash)
	if !match {
		return LoginResponse{}, apperror.NewUnauthorized("Invalid credentials")
	}

	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey: s.jwtAccessKey,
		issuer:     s.issuer,
		userID:     ulidutil.MustFromBytes(user.ID),
		expiry:     s.accessTokenExpiry,
	})
	if err != nil {
		return LoginResponse{}, apperror.NewInternalServerError("Token generation error")
	}

	userID := ulidutil.MustFromBytes(user.ID)
	refreshTokenModel := model.RefreshTokens{
		ID:       ulid.Make().Bytes(),
		UserID:   userID.Bytes(),
		ParentID: nil,
		IssuedAt: time.Now(),
		// TODO refactor this so we don't have the magic number everywhere
		ExpiresAt: time.Now().Add(time.Duration(168) * time.Hour),
		RevokedAt: nil,
		IPAddress: ip,
		UserAgent: userAgent,
	}
	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return LoginResponse{}, err
	}

	refreshToken, err := GenerateRefreshToken(GenerateRefreshTokenParams{
		privateKey: s.jwtRefreshKey,
		issuer:     s.issuer,
		userID:     userID,
		tokenID:    ulidutil.MustFromBytes(refreshTokenModel.ID),
		expiry:     s.refreshTokenExpiry,
	})
	if err != nil {
		return LoginResponse{}, apperror.NewInternalServerError("Token generation error")
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
		RefreshToken: refreshToken,
	}

	return response, nil
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

func (s *AuthService) Refresh(params RefreshParams, ip string, userAgent string) (RefreshResponse, error) {
	_, claims, err := ValidateToken(s.jwtRefreshKey.Public().(ed25519.PublicKey), params.RefreshToken)
	if err != nil {
		return RefreshResponse{}, err
	}

	tokenID, err := ulid.Parse(claims.ID)
	if err != nil {
		return RefreshResponse{}, apperror.NewBadRequest("Invalid token ID format")
	}

	refreshToken, err := s.refreshTokenRepo.GetByID(tokenID)
	if err != nil {
		return RefreshResponse{}, err
	}

	if refreshToken == nil {
		return RefreshResponse{}, apperror.NewUnauthorized("Invalid token")
	}

	if refreshToken.RevokedAt != nil {
		return RefreshResponse{}, apperror.NewUnauthorized("Invalid token")
	}

	refreshTokenULID := ulidutil.MustFromBytes(refreshToken.ID)
	if err := s.refreshTokenRepo.Revoke(refreshTokenULID); err != nil {
		return RefreshResponse{}, err
	}

	userID := ulidutil.MustFromBytes(refreshToken.UserID)
	accessToken, err := GenerateAccessToken(GenerateAccessTokenParams{
		privateKey: s.jwtAccessKey,
		issuer:     s.issuer,
		userID:     userID,
		expiry:     s.accessTokenExpiry,
	})
	if err != nil {
		return RefreshResponse{}, apperror.NewInternalServerError("Token generation error")
	}

	newRefreshTokenID := ulid.Make()
	newRefreshTokenModel := model.RefreshTokens{
		ID:       newRefreshTokenID.Bytes(),
		UserID:   refreshToken.UserID,
		ParentID: &refreshToken.ID,
		IssuedAt: time.Now(),
		// TODO refactor this so we don't have the magic number everywhere
		ExpiresAt: time.Now().Add(time.Duration(168) * time.Hour),
		RevokedAt: nil,
		IPAddress: ip,
		UserAgent: userAgent,
	}
	if err := s.refreshTokenRepo.Create(newRefreshTokenModel); err != nil {
		return RefreshResponse{}, err
	}

	newRefreshToken, err := GenerateRefreshToken(GenerateRefreshTokenParams{
		privateKey: s.jwtRefreshKey,
		issuer:     s.issuer,
		userID:     userID,
		tokenID:    ulidutil.MustFromBytes(newRefreshTokenModel.ID),
		expiry:     s.refreshTokenExpiry,
	})

	if err != nil {
		return RefreshResponse{}, apperror.NewInternalServerError("Token generation error")
	}

	return RefreshResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
		RefreshToken: newRefreshToken,
	}, nil
}
