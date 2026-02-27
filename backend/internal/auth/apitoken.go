package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateAPIToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "bpt_" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func APITokenPrefix(token string) string {
	if len(token) <= 6 {
		return token
	}
	return token[:6]
}

// HashAPIToken computes HMAC-SHA256(token, secret) for deterministic secure lookup.
func HashAPIToken(secret, token string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}
