package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

type JWTManager struct {
	signingKey []byte
	ttl        time.Duration
}

func NewJWTManager(signingKey string, ttl time.Duration) *JWTManager {
	return &JWTManager{signingKey: []byte(signingKey), ttl: ttl}
}

func (m *JWTManager) GenerateAccessToken(userID int64, username string, now time.Time) (string, int64, error) {
	expiresAt := now.Add(m.ttl)
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(m.ttl.Seconds()), nil
}

func (m *JWTManager) ParseAccessToken(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrTokenExpired
		}
		return Claims{}, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return Claims{}, ErrTokenInvalid
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return Claims{}, ErrTokenInvalid
	}
	username, ok := claims["username"].(string)
	if !ok {
		return Claims{}, ErrTokenInvalid
	}
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return Claims{}, ErrTokenInvalid
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return Claims{}, ErrTokenInvalid
	}

	return Claims{
		UserID:    int64(sub),
		Username:  username,
		IssuedAt:  time.Unix(int64(iatFloat), 0),
		ExpiresAt: time.Unix(int64(expFloat), 0),
	}, nil
}
