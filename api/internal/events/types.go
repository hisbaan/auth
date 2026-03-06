package events

type EventType string

const (
	UserCreated                  EventType = "user.created"
	UserDeleted                  EventType = "user.deleted"
	UserUpdated                  EventType = "user.updated"
	UserEmailVerified            EventType = "user.email_verified"
	UserEmailVerificationCreated EventType = "user.email_verification_created"

	APIKeyCreated EventType = "api_key.created"
	APIKeyRevoked EventType = "api_key.revoked"
	APIKeyUsed    EventType = "api_key.used"

	AuthenticationPasswordSucceeded EventType = "authentication.password_succeeded"
	AuthenticationPasswordFailed    EventType = "authentication.password_failed"

	RoleCreated    EventType = "role.created"
	RoleDeleted    EventType = "role.deleted"
	RoleAssigned   EventType = "role.assigned"
	RoleUnassigned EventType = "role.unassigned"

	PasswordResetCreated   EventType = "password_reset.created"
	PasswordResetSucceeded EventType = "password_reset.succeeded"

	AccessTokenCreated  EventType = "access_token.created"
	RefreshTokenCreated EventType = "refresh_token.created"
	RefreshTokenRevoked EventType = "refresh_token.revoked"
)

type UserCreatedData struct {
	UserID string `json:"user_id"`
}

type UserDeletedData struct {
	UserID string `json:"user_id"`
}

type UserUpdatedData struct {
	UserID string `json:"user_id"`
}

type UserEmailVerifiedData struct {
	UserID string `json:"user_id"`
}

type UserEmailVerificationCreatedData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
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
	UserID string `json:"user_id"`
}

type AuthenticationPasswordFailedData struct {
	Email string `json:"email"`
}

type RoleCreatedData struct {
	Role string `json:"role"`
}

type RoleDeletedData struct {
	Role string `json:"role"`
}

type RoleAssignedData struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type RoleUnassignedData struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type PasswordResetCreatedData struct {
	UserID               string `json:"user_id"`
	PasswordResetTokenID string `json:"password_reset_token_id"`
}

type PasswordResetSucceededData struct {
	UserID               string `json:"user_id"`
	PasswordResetTokenID string `json:"password_reset_token_id"`
}

type AccessTokenCreatedData struct {
	UserID string `json:"user_id"`
}

type RefreshTokenCreatedData struct {
	UserID         string `json:"user_id"`
	RefreshTokenID string `json:"refresh_token_id"`
}

type RefreshTokenRevokedData struct {
	UserID         string `json:"user_id"`
	RefreshTokenID string `json:"refresh_token_id"`
}
