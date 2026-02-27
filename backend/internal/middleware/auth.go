package middleware

import (
	"strings"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			apperror.Write(c, apperror.Unauthorized("Missing authorization token"))
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseAccessToken(parts[1])
		if err != nil {
			if err == auth.ErrTokenExpired {
				apperror.Write(c, apperror.Expired("Access token expired"))
			} else {
				apperror.Write(c, apperror.Unauthorized("Invalid access token"))
			}
			c.Abort()
			return
		}

		c.Set(auth.ClaimsContextKey, claims)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (auth.Claims, bool) {
	value, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		return auth.Claims{}, false
	}
	claims, ok := value.(auth.Claims)
	if !ok {
		return auth.Claims{}, false
	}
	return claims, true
}
