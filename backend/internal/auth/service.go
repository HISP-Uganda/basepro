package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/audit"
)

type Service struct {
	repo         Repository
	auditService *audit.Service
	jwt          *JWTManager
	accessTTL    time.Duration
	refreshTTL   time.Duration
	now          func() time.Time
}

func NewService(repo Repository, auditService *audit.Service, jwt *JWTManager, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:         repo,
		auditService: auditService,
		jwt:          jwt,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Login(ctx context.Context, username, password, ip, userAgent string) (AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil || !user.IsActive || ComparePassword(user.PasswordHash, password) != nil {
		s.logAudit(ctx, audit.Event{
			Action:     "auth.login.failure",
			EntityType: "auth",
			Metadata: map[string]any{
				"username":   username,
				"reason":     "invalid_credentials",
				"ip":         ip,
				"user_agent": userAgent,
			},
		})
		return AuthResponse{}, apperror.Unauthorized("Invalid credentials")
	}

	response, err := s.issueTokens(ctx, user.ID, user.Username)
	if err != nil {
		return AuthResponse{}, err
	}

	s.logAudit(ctx, audit.Event{
		Action:      "auth.login.success",
		ActorUserID: &user.ID,
		EntityType:  "auth",
		EntityID:    strPtr(user.Username),
		Metadata: map[string]any{
			"ip":         ip,
			"user_agent": userAgent,
		},
	})

	return response, nil
}

func (s *Service) Refresh(ctx context.Context, presentedToken, ip, userAgent string) (AuthResponse, error) {
	now := s.now()
	hash := HashToken(presentedToken)
	token, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		s.logAudit(ctx, audit.Event{
			Action:     "auth.refresh.failure",
			EntityType: "auth",
			Metadata: map[string]any{
				"reason":     "invalid_refresh_token",
				"ip":         ip,
				"user_agent": userAgent,
			},
		})
		return AuthResponse{}, apperror.RefreshInvalid("Refresh token is invalid")
	}

	if token.RevokedAt != nil {
		_ = s.repo.RevokeAllActiveRefreshTokensForUser(ctx, token.UserID, now)
		s.logAudit(ctx, audit.Event{
			Action:      "auth.refresh.reused",
			ActorUserID: &token.UserID,
			EntityType:  "auth",
			Metadata: map[string]any{
				"ip":         ip,
				"user_agent": userAgent,
			},
		})
		return AuthResponse{}, apperror.RefreshReused("Refresh token has been reused")
	}

	if now.After(token.ExpiresAt) {
		s.logAudit(ctx, audit.Event{
			Action:      "auth.refresh.failure",
			ActorUserID: &token.UserID,
			EntityType:  "auth",
			Metadata: map[string]any{
				"reason":     "expired_refresh_token",
				"ip":         ip,
				"user_agent": userAgent,
			},
		})
		return AuthResponse{}, apperror.RefreshInvalid("Refresh token is invalid")
	}

	newPlain, err := GenerateRefreshToken()
	if err != nil {
		return AuthResponse{}, errors.New("failed to generate refresh token")
	}
	newHash := HashToken(newPlain)

	newRecord, err := s.repo.CreateRefreshToken(ctx, RefreshToken{
		UserID:    token.UserID,
		TokenHash: newHash,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	if err := s.repo.RevokeRefreshToken(ctx, token.ID, &newRecord.ID, now); err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return AuthResponse{}, apperror.Unauthorized("Invalid credentials")
	}

	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(token.UserID, user.Username, now)
	if err != nil {
		return AuthResponse{}, err
	}

	s.logAudit(ctx, audit.Event{
		Action:      "auth.refresh.success",
		ActorUserID: &token.UserID,
		EntityType:  "auth",
		Metadata: map[string]any{
			"ip":         ip,
			"user_agent": userAgent,
		},
	})

	return AuthResponse{AccessToken: accessToken, RefreshToken: newPlain, ExpiresIn: expiresIn}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken, authHeader, ip, userAgent string) error {
	now := s.now()
	var actor *int64

	if refreshToken != "" {
		hash := HashToken(refreshToken)
		token, err := s.repo.GetRefreshTokenByHash(ctx, hash)
		if err == nil {
			actor = &token.UserID
			_ = s.repo.RevokeRefreshToken(ctx, token.ID, nil, now)
		}
	} else if claims, ok := s.parseBearer(authHeader); ok {
		actor = &claims.UserID
		_ = s.repo.RevokeAllActiveRefreshTokensForUser(ctx, claims.UserID, now)
	}

	s.logAudit(ctx, audit.Event{
		Action:      "auth.logout",
		ActorUserID: actor,
		EntityType:  "auth",
		Metadata: map[string]any{
			"ip":         ip,
			"user_agent": userAgent,
		},
	})

	return nil
}

func (s *Service) Me(claims Claims) map[string]any {
	return map[string]any{
		"id":       claims.UserID,
		"username": claims.Username,
		"roles":    []string{},
	}
}

func (s *Service) issueTokens(ctx context.Context, userID int64, username string) (AuthResponse, error) {
	now := s.now()
	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(userID, username, now)
	if err != nil {
		return AuthResponse{}, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return AuthResponse{}, err
	}

	_, err = s.repo.CreateRefreshToken(ctx, RefreshToken{
		UserID:    userID,
		TokenHash: HashToken(refreshToken),
		IssuedAt:  now,
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: expiresIn}, nil
}

func (s *Service) parseBearer(header string) (Claims, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return Claims{}, false
	}
	claims, err := s.jwt.ParseAccessToken(parts[1])
	if err != nil {
		return Claims{}, false
	}
	return claims, true
}

func (s *Service) logAudit(ctx context.Context, event audit.Event) {
	if s.auditService == nil {
		return
	}
	_ = s.auditService.Log(ctx, event)
}

func strPtr(value string) *string {
	return &value
}
