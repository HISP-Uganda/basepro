package users

import (
	"net/http"
	"strconv"
	"strings"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/auth"
	"basepro/backend/internal/listquery"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createUserRequest struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	Email          *string  `json:"email"`
	Language       *string  `json:"language"`
	FirstName      *string  `json:"firstName"`
	LastName       *string  `json:"lastName"`
	DisplayName    *string  `json:"displayName"`
	PhoneNumber    *string  `json:"phoneNumber"`
	WhatsappNumber *string  `json:"whatsappNumber"`
	TelegramHandle *string  `json:"telegramHandle"`
	IsActive       *bool    `json:"isActive"`
	Roles          []string `json:"roles"`
}

type updateUserRequest struct {
	Username       *string   `json:"username"`
	Password       *string   `json:"password"`
	Email          *string   `json:"email"`
	Language       *string   `json:"language"`
	FirstName      *string   `json:"firstName"`
	LastName       *string   `json:"lastName"`
	DisplayName    *string   `json:"displayName"`
	PhoneNumber    *string   `json:"phoneNumber"`
	WhatsappNumber *string   `json:"whatsappNumber"`
	TelegramHandle *string   `json:"telegramHandle"`
	IsActive       *bool     `json:"isActive"`
	Roles          *[]string `json:"roles"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (h *Handler) List(c *gin.Context) {
	page, err := listquery.ParseInt(c.Query("page"), 1, 1, 100000, "page")
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"page": []string{err.Error()}}))
		return
	}
	pageSize, err := listquery.ParseInt(c.Query("pageSize"), 25, 1, 200, "pageSize")
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"pageSize": []string{err.Error()}}))
		return
	}
	sortField, sortOrder, err := listquery.ParseSort(c.Query("sort"))
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"sort": []string{err.Error()}}))
		return
	}
	filterField, filterValue, err := listquery.ParseFilter(c.Query("filter"))
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"filter": []string{err.Error()}}))
		return
	}
	search := listquery.ResolveSearch(c.Query("q"), filterField, filterValue, map[string]struct{}{
		"username":    {},
		"email":       {},
		"displayName": {},
	})

	list, err := h.service.ListUsers(c.Request.Context(), ListQuery{
		Page:      page,
		PageSize:  pageSize,
		SortField: sortField,
		SortOrder: sortOrder,
		Filter:    search,
	})
	if err != nil {
		apperror.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":      list.Items,
		"totalCount": list.Total,
		"page":       list.Page,
		"pageSize":   list.PageSize,
	})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"invalid user id"}}))
		return
	}

	item, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Create(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Unauthorized"))
		return
	}

	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"invalid JSON payload"}}))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	created, err := h.service.CreateUser(c.Request.Context(), CreateInput{
		Username:       req.Username,
		Password:       req.Password,
		Email:          req.Email,
		Language:       req.Language,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DisplayName:    req.DisplayName,
		PhoneNumber:    req.PhoneNumber,
		WhatsappNumber: req.WhatsappNumber,
		TelegramHandle: req.TelegramHandle,
		IsActive:       isActive,
		Roles:          normalizeRoles(req.Roles),
		ActorID:        actorUserID(principal),
	})
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) Patch(c *gin.Context) {
	h.update(c)
}

func (h *Handler) Put(c *gin.Context) {
	h.update(c)
}

func (h *Handler) update(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Unauthorized"))
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"invalid user id"}}))
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"invalid JSON payload"}}))
		return
	}
	if req.Username == nil && req.Password == nil && req.Email == nil && req.Language == nil && req.FirstName == nil && req.LastName == nil && req.DisplayName == nil && req.PhoneNumber == nil && req.WhatsappNumber == nil && req.TelegramHandle == nil && req.Roles == nil && req.IsActive == nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"at least one update field is required"}}))
		return
	}

	var roles *[]string
	if req.Roles != nil {
		cleaned := normalizeRoles(*req.Roles)
		roles = &cleaned
	}

	updated, updateErr := h.service.UpdateUser(c.Request.Context(), UpdateInput{
		UserID:         id,
		Username:       req.Username,
		Password:       req.Password,
		Email:          req.Email,
		Language:       req.Language,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DisplayName:    req.DisplayName,
		PhoneNumber:    req.PhoneNumber,
		WhatsappNumber: req.WhatsappNumber,
		TelegramHandle: req.TelegramHandle,
		Roles:          roles,
		IsActive:       req.IsActive,
		ActorID:        actorUserID(principal),
	})
	if updateErr != nil {
		apperror.Write(c, updateErr)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		apperror.Write(c, apperror.Unauthorized("Unauthorized"))
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"invalid user id"}}))
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Password) == "" {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"password": []string{"is required"}}))
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), actorUserID(principal), id, req.Password); err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

func normalizeRoles(roles []string) []string {
	if len(roles) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		trimmed := strings.TrimSpace(role)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
