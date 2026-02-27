package auth

import (
	"net/http"

	"basepro/backend/internal/apperror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		apperror.Write(c, apperror.Unauthorized("Invalid credentials"))
		return
	}

	response, err := h.service.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		apperror.Write(c, apperror.RefreshInvalid("Refresh token is invalid"))
		return
	}

	response, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken, c.GetHeader("Authorization"), c.ClientIP(), c.Request.UserAgent()); err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Me(c *gin.Context) {
	value, ok := c.Get(ClaimsContextKey)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Missing authorization token"))
		return
	}
	claims, ok := value.(Claims)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Invalid access token"))
		return
	}
	c.JSON(http.StatusOK, h.service.Me(claims))
}
