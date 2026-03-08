package bootstrap

import (
	"errors"
	"net/http"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(c *gin.Context) {
	if h.service == nil {
		apperror.Write(c, errors.New("bootstrap service is not configured"))
		return
	}

	var principal *auth.Principal
	if resolved, ok := principalFromContext(c); ok {
		principal = &resolved
	}

	response, err := h.service.Build(c.Request.Context(), principal)
	if err != nil {
		apperror.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func principalFromContext(c *gin.Context) (auth.Principal, bool) {
	value, ok := c.Get(auth.PrincipalContextKey)
	if !ok {
		return auth.Principal{}, false
	}
	principal, ok := value.(auth.Principal)
	return principal, ok
}
