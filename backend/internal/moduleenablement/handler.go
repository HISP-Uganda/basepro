package moduleenablement

import (
	"errors"
	"net/http"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service            *Service
	getConfigOverrides func() map[string]bool
}

func NewHandler(service *Service, getConfigOverrides func() map[string]bool) *Handler {
	if getConfigOverrides == nil {
		getConfigOverrides = func() map[string]bool { return nil }
	}
	return &Handler{
		service:            service,
		getConfigOverrides: getConfigOverrides,
	}
}

func (h *Handler) GetEffective(c *gin.Context) {
	if h.service == nil {
		apperror.Write(c, errors.New("module enablement service is not configured"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"modules": h.service.ListEffective(h.getConfigOverrides()),
	})
}

type updateRuntimeOverridesRequest struct {
	Modules []RuntimeModuleOverride `json:"modules"`
}

func (h *Handler) UpdateRuntimeOverrides(c *gin.Context) {
	if h.service == nil {
		apperror.Write(c, errors.New("module enablement service is not configured"))
		return
	}
	principal, ok := principalFromContext(c)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Unauthorized"))
		return
	}

	var req updateRuntimeOverridesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{
			"body": []string{"invalid JSON payload"},
		}))
		return
	}

	modules, err := h.service.UpdateRuntimeOverrides(
		c.Request.Context(),
		req.Modules,
		h.getConfigOverrides(),
		actorUserID(principal),
	)
	if err != nil {
		apperror.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
	})
}

func principalFromContext(c *gin.Context) (auth.Principal, bool) {
	value, ok := c.Get(auth.PrincipalContextKey)
	if !ok {
		return auth.Principal{}, false
	}
	principal, ok := value.(auth.Principal)
	return principal, ok
}

func actorUserID(principal auth.Principal) *int64 {
	if principal.Type != "user" {
		return nil
	}
	return &principal.UserID
}
