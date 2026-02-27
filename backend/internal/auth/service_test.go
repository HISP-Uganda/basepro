package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/audit"
)

type fakeRepo struct {
	usersByID       map[int64]*User
	usersByUsername map[string]*User
	tokensByHash    map[string]*RefreshToken
	tokensByID      map[int64]*RefreshToken
	nextTokenID     int64
}

func newFakeRepo(user *User) *fakeRepo {
	return &fakeRepo{
		usersByID:       map[int64]*User{user.ID: user},
		usersByUsername: map[string]*User{user.Username: user},
		tokensByHash:    map[string]*RefreshToken{},
		tokensByID:      map[int64]*RefreshToken{},
		nextTokenID:     1,
	}
}

func (r *fakeRepo) GetUserByUsername(_ context.Context, username string) (*User, error) {
	user, ok := r.usersByUsername[username]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (r *fakeRepo) GetUserByID(_ context.Context, userID int64) (*User, error) {
	user, ok := r.usersByID[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (r *fakeRepo) GetRefreshTokenByHash(_ context.Context, hash string) (*RefreshToken, error) {
	token, ok := r.tokensByHash[hash]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *token
	return &copy, nil
}

func (r *fakeRepo) CreateRefreshToken(_ context.Context, token RefreshToken) (*RefreshToken, error) {
	token.ID = r.nextTokenID
	r.nextTokenID++
	copy := token
	r.tokensByHash[token.TokenHash] = &copy
	r.tokensByID[token.ID] = &copy
	return &copy, nil
}

func (r *fakeRepo) RevokeRefreshToken(_ context.Context, tokenID int64, replacedByTokenID *int64, now time.Time) error {
	token, ok := r.tokensByID[tokenID]
	if !ok {
		return ErrNotFound
	}
	if token.RevokedAt == nil {
		token.RevokedAt = &now
	}
	token.ReplacedByTokenID = replacedByTokenID
	token.UpdatedAt = now
	return nil
}

func (r *fakeRepo) RevokeAllActiveRefreshTokensForUser(_ context.Context, userID int64, now time.Time) error {
	for _, token := range r.tokensByID {
		if token.UserID == userID && token.RevokedAt == nil {
			t := now
			token.RevokedAt = &t
			token.UpdatedAt = now
		}
	}
	return nil
}

type fakeAuditRepo struct {
	events []audit.Event
}

func (r *fakeAuditRepo) Insert(_ context.Context, event audit.Event) error {
	r.events = append(r.events, event)
	return nil
}

func TestLoginSuccessReturnsTokensAndStoresRefreshHash(t *testing.T) {
	passwordHash, err := HashPassword("secret", 4)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &User{ID: 10, Username: "alice", PasswordHash: passwordHash, IsActive: true}
	repo := newFakeRepo(user)
	auditRepo := &fakeAuditRepo{}
	service := NewService(repo, audit.NewService(auditRepo), NewJWTManager("test-key", 5*time.Minute), 5*time.Minute, 24*time.Hour)

	resp, err := service.Login(context.Background(), "alice", "secret", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if resp.ExpiresIn <= 0 {
		t.Fatalf("expected positive expiresIn, got %d", resp.ExpiresIn)
	}

	hash := HashToken(resp.RefreshToken)
	stored, err := repo.GetRefreshTokenByHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("stored refresh token not found: %v", err)
	}
	if stored.TokenHash == resp.RefreshToken {
		t.Fatal("refresh token must be stored as hash, not plaintext")
	}
}

func TestLoginFailureReturnsUnauthorized(t *testing.T) {
	passwordHash, err := HashPassword("secret", 4)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &User{ID: 1, Username: "alice", PasswordHash: passwordHash, IsActive: true}
	repo := newFakeRepo(user)
	service := NewService(repo, audit.NewService(&fakeAuditRepo{}), NewJWTManager("test-key", time.Minute), time.Minute, 2*time.Hour)

	_, err = service.Login(context.Background(), "alice", "wrong", "127.0.0.1", "test-agent")
	if err == nil {
		t.Fatal("expected login error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperror.CodeAuthUnauthorized {
		t.Fatalf("expected %s, got %s", apperror.CodeAuthUnauthorized, appErr.Code)
	}
}

func TestRefreshSuccessRotatesTokenAndNewTokenWorks(t *testing.T) {
	passwordHash, err := HashPassword("secret", 4)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &User{ID: 2, Username: "bob", PasswordHash: passwordHash, IsActive: true}
	repo := newFakeRepo(user)
	service := NewService(repo, audit.NewService(&fakeAuditRepo{}), NewJWTManager("test-key", 5*time.Minute), 5*time.Minute, 24*time.Hour)

	loginResp, err := service.Login(context.Background(), "bob", "secret", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	oldRecord, err := repo.GetRefreshTokenByHash(context.Background(), HashToken(loginResp.RefreshToken))
	if err != nil {
		t.Fatalf("get old record: %v", err)
	}

	refreshResp, err := service.Refresh(context.Background(), loginResp.RefreshToken, "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	updatedOld, err := repo.GetRefreshTokenByHash(context.Background(), HashToken(loginResp.RefreshToken))
	if err != nil {
		t.Fatalf("get updated old record: %v", err)
	}
	if updatedOld.RevokedAt == nil {
		t.Fatal("expected old token to be revoked")
	}
	if updatedOld.ReplacedByTokenID == nil || *updatedOld.ReplacedByTokenID == oldRecord.ID {
		t.Fatal("expected old token replaced_by_token_id to point to new token")
	}

	if _, err := service.Refresh(context.Background(), refreshResp.RefreshToken, "127.0.0.1", "agent"); err != nil {
		t.Fatalf("newly issued refresh token should work: %v", err)
	}
}

func TestRefreshReuseDetectionRevokesActiveTokens(t *testing.T) {
	passwordHash, err := HashPassword("secret", 4)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &User{ID: 3, Username: "carol", PasswordHash: passwordHash, IsActive: true}
	repo := newFakeRepo(user)
	service := NewService(repo, audit.NewService(&fakeAuditRepo{}), NewJWTManager("test-key", 5*time.Minute), 5*time.Minute, 24*time.Hour)

	loginResp, err := service.Login(context.Background(), "carol", "secret", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	rotatedResp, err := service.Refresh(context.Background(), loginResp.RefreshToken, "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	_, err = service.Refresh(context.Background(), loginResp.RefreshToken, "127.0.0.1", "agent")
	if err == nil {
		t.Fatal("expected reuse detection error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperror.CodeAuthRefreshReuse {
		t.Fatalf("expected %s, got %s", apperror.CodeAuthRefreshReuse, appErr.Code)
	}

	active, err := repo.GetRefreshTokenByHash(context.Background(), HashToken(rotatedResp.RefreshToken))
	if err != nil {
		t.Fatalf("lookup active token: %v", err)
	}
	if active.RevokedAt == nil {
		t.Fatal("expected all active user tokens to be revoked on reuse detection")
	}
}
