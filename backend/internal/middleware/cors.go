package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS allows configured origins, including simple wildcard suffixes like
// "wails://wails.localhost:*" for Wails runtime random ports.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	normalized := make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if allowedOrigin, ok := matchAllowedOrigin(origin, normalized); ok {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", allowedOrigin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-API-Token")
			h.Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func matchAllowedOrigin(origin string, allowedOrigins []string) (string, bool) {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return "*", true
		}

		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return origin, true
			}
			continue
		}

		if origin == allowed {
			return origin, true
		}
	}

	return "", false
}
