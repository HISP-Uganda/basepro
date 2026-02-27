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

		principal := auth.Principal{
			Type:        "user",
			UserID:      claims.UserID,
			Username:    claims.Username,
			Permissions: []string{},
			IsAdmin:     claims.UserID == 1,
		}

		c.Set(auth.ClaimsContextKey, claims)
		c.Set(auth.PrincipalContextKey, principal)
		c.Next()
	}
}

func APITokenAuth(service *auth.Service, headerName string, allowBearer bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader(headerName))
		if token == "" && allowBearer {
			parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				candidate := strings.TrimSpace(parts[1])
				// JWTs are handled by JWTAuth middleware on protected JWT routes.
				if strings.Count(candidate, ".") != 2 {
					token = candidate
				}
			}
		}

		if token == "" {
			c.Next()
			return
		}

		principal, err := service.AuthenticateAPIToken(c.Request.Context(), token, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			apperror.Write(c, err)
			c.Abort()
			return
		}

		c.Set(auth.PrincipalContextKey, principal)
		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := PrincipalFromContext(c)
		if !ok {
			apperror.Write(c, apperror.Unauthorized("Unauthorized"))
			c.Abort()
			return
		}

		if principal.Type == "user" {
			if principal.IsAdmin {
				c.Next()
				return
			}
			apperror.Write(c, apperror.Unauthorized("Permission denied"))
			c.Abort()
			return
		}

		for _, candidate := range principal.Permissions {
			if candidate == permission {
				c.Next()
				return
			}
		}

		apperror.Write(c, apperror.Unauthorized("Permission denied"))
		c.Abort()
	}
}

func RequireAdminUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := PrincipalFromContext(c)
		if !ok || principal.Type != "user" || !principal.IsAdmin {
			apperror.Write(c, apperror.Unauthorized("Admin access required"))
			c.Abort()
			return
		}
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

func PrincipalFromContext(c *gin.Context) (auth.Principal, bool) {
	value, ok := c.Get(auth.PrincipalContextKey)
	if !ok {
		return auth.Principal{}, false
	}
	principal, ok := value.(auth.Principal)
	return principal, ok
}
