package auth

import "time"

const ClaimsContextKey = "auth_claims"

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	IsActive     bool   `db:"is_active"`
}

type RefreshToken struct {
	ID                int64      `db:"id"`
	UserID            int64      `db:"user_id"`
	TokenHash         string     `db:"token_hash"`
	IssuedAt          time.Time  `db:"issued_at"`
	ExpiresAt         time.Time  `db:"expires_at"`
	RevokedAt         *time.Time `db:"revoked_at"`
	ReplacedByTokenID *int64     `db:"replaced_by_token_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type Claims struct {
	UserID    int64
	Username  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}
