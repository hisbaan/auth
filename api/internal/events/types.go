package events

type EventType string

type EventReason string

const (
	UserCreated                  EventType = "user.created"
	UserDeleted                  EventType = "user.deleted"
	UserUpdated                  EventType = "user.updated"
	UserPasswordChanged          EventType = "user.password_changed"
	UserEmailVerified            EventType = "user.email_verified"
	UserEmailVerificationCreated EventType = "user.email_verification_created"
	UserEmailVerificationRevoked EventType = "user.email_verification_revoked"

	APIKeyCreated EventType = "api_key.created"
	APIKeyRevoked EventType = "api_key.revoked"
	APIKeyUsed    EventType = "api_key.used"

	AuthenticationPasswordSucceeded   EventType = "authentication.password_succeeded"
	AuthenticationPasswordFailed      EventType = "authentication.password_failed"
	AuthenticationLogout              EventType = "authentication.logout"
	AuthenticationRefreshFailed       EventType = "authentication.refresh_failed"
	AuthenticationRefreshTokenRotated EventType = "authentication.refresh_token_rotated"
	AuthenticationVerifyEmailFailed   EventType = "authentication.verify_email_failed"

	RoleCreated    EventType = "role.created"
	RoleDeleted    EventType = "role.deleted"
	RoleAssigned   EventType = "role.assigned"
	RoleUnassigned EventType = "role.unassigned"

	PasswordResetCreated   EventType = "password_reset.created"
	PasswordResetSucceeded EventType = "password_reset.succeeded"
	PasswordResetFailed    EventType = "password_reset.failed"

	AccessTokenCreated  EventType = "access_token.created"
	RefreshTokenCreated EventType = "refresh_token.created"
	RefreshTokenRevoked EventType = "refresh_token.revoked"
)

const (
	EventReasonInvalidToken        EventReason = "invalid_token"
	EventReasonExpiredToken        EventReason = "expired_token"
	EventReasonRevokedToken        EventReason = "revoked_token"
	EventReasonInvalidSignature    EventReason = "invalid_signature"
	EventReasonUnknownRefreshToken EventReason = "unknown_refresh_token"
)

type UserCreatedData struct {
}

type UserDeletedData struct {
}

type UserUpdatedData struct {
}

type UserPasswordChangedData struct {
}

type UserEmailVerifiedData struct {
}

type UserEmailVerificationCreatedData struct {
	Email string `json:"email"`
}

type UserEmailVerificationRevokedData struct {
	Email string `json:"email"`
}

type APIKeyCreatedData struct {
	UserID   string `json:"user_id"`
	APIKeyID string `json:"api_key_id"`
}

type APIKeyRevokedData struct {
	UserID   string `json:"user_id"`
	APIKeyID string `json:"api_key_id"`
}

type APIKeyUsedData struct {
	UserID   string `json:"user_id"`
	APIKeyID string `json:"api_key_id"`
}

type AuthenticationPasswordSucceededData struct {
}

type AuthenticationPasswordFailedData struct {
	Email string `json:"email"`
}

type AuthenticationLogoutData struct {
}

type AuthenticationRefreshFailedData struct {
	RefreshTokenID string      `json:"refresh_token_id"`
	Reason         EventReason `json:"reason"`
}

type AuthenticationRefreshTokenRotatedData struct {
	OldRefreshTokenID string `json:"old_refresh_token_id"`
	NewRefreshTokenID string `json:"new_refresh_token_id"`
}

type AuthenticationVerifyEmailFailedData struct {
	Reason EventReason `json:"reason"`
}

type RoleCreatedData struct {
	Role    string `json:"role"`
	ActorID string `json:"actor_id"`
}

type RoleDeletedData struct {
	Role    string `json:"role"`
	ActorID string `json:"actor_id"`
}

type RoleAssignedData struct {
	Role    string `json:"role"`
	ActorID string `json:"actor_id"`
}

type RoleUnassignedData struct {
	Role    string `json:"role"`
	ActorID string `json:"actor_id"`
}

type PasswordResetCreatedData struct {
	PasswordResetTokenID string `json:"password_reset_token_id"`
}

type PasswordResetSucceededData struct {
	PasswordResetTokenID string `json:"password_reset_token_id"`
}

type PasswordResetFailedData struct {
	PasswordResetTokenID string      `json:"password_reset_token_id"`
	Reason               EventReason `json:"reason"`
}

type AccessTokenCreatedData struct {
}

type RefreshTokenCreatedData struct {
	RefreshTokenID string `json:"refresh_token_id"`
}

type RefreshTokenRevokedData struct {
	RefreshTokenID string `json:"refresh_token_id"`
}
