package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	CreateRefreshToken(ctx context.Context, token RefreshToken) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID int64, replacedByTokenID *int64, now time.Time) error
	RevokeAllActiveRefreshTokensForUser(ctx context.Context, userID int64, now time.Time) error
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password_hash, is_active
		FROM users
		WHERE username = $1
	`, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &user, nil
}

func (r *SQLRepository) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password_hash, is_active
		FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *SQLRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.GetContext(ctx, &token, `
		SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by_token_id, created_at, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &token, nil
}

func (r *SQLRepository) CreateRefreshToken(ctx context.Context, token RefreshToken) (*RefreshToken, error) {
	var created RefreshToken
	err := r.db.GetContext(ctx, &created, `
		INSERT INTO refresh_tokens (user_id, token_hash, issued_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by_token_id, created_at, updated_at
	`, token.UserID, token.TokenHash, token.IssuedAt, token.ExpiresAt, token.CreatedAt, token.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return &created, nil
}

func (r *SQLRepository) RevokeRefreshToken(ctx context.Context, tokenID int64, replacedByTokenID *int64, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, $2),
		    replaced_by_token_id = COALESCE($3, replaced_by_token_id),
		    updated_at = $2
		WHERE id = $1
	`, tokenID, now, replacedByTokenID)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *SQLRepository) RevokeAllActiveRefreshTokensForUser(ctx context.Context, userID int64, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, $2), updated_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, now)
	if err != nil {
		return fmt.Errorf("revoke active refresh tokens: %w", err)
	}
	return nil
}
