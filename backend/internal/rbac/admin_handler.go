package rbac

import (
	"net/http"
	"strconv"
	"strings"

	"basepro/backend/internal/apperror"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service *AdminService
}

func NewAdminHandler(service *AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type updateRoleRequest struct {
	Name        *string   `json:"name"`
	Permissions *[]string `json:"permissions"`
}

func (h *AdminHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "25"))
	sortField, sortOrder := parseSortQuery(c.Query("sort"))
	filterField, filterValue := parseFilterQuery(c.Query("filter"))
	filter := filterForRoles(filterField, filterValue)

	result, err := h.service.ListRoles(c.Request.Context(), RoleListQuery{
		Page:      page,
		PageSize:  pageSize,
		SortField: sortField,
		SortOrder: sortOrder,
		Filter:    filter,
	})
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      result.Items,
		"totalCount": result.Total,
		"page":       result.Page,
		"pageSize":   result.PageSize,
	})
}

func (h *AdminHandler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"invalid JSON payload"}}))
		return
	}

	created, err := h.service.CreateRole(c.Request.Context(), RoleCreateInput{
		Name:        req.Name,
		Permissions: req.Permissions,
		ActorUserID: actorUserIDFromContext(c),
	})
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *AdminHandler) GetRole(c *gin.Context) {
	id, err := parseRoleIDParam(c.Param("id"))
	if err != nil {
		apperror.Write(c, err)
		return
	}

	includeUsers := parseBoolQuery(c.Query("includeUsers"))
	detail, err := h.service.GetRoleDetail(c.Request.Context(), id, includeUsers)
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *AdminHandler) PatchRole(c *gin.Context) {
	h.updateRole(c)
}

func (h *AdminHandler) PutRole(c *gin.Context) {
	h.updateRole(c)
}

func (h *AdminHandler) updateRole(c *gin.Context) {
	id, err := parseRoleIDParam(c.Param("id"))
	if err != nil {
		apperror.Write(c, err)
		return
	}

	var req updateRoleRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"invalid JSON payload"}}))
		return
	}
	if req.Name == nil && req.Permissions == nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"at least one update field is required"}}))
		return
	}

	updated, updateErr := h.service.UpdateRole(c.Request.Context(), RoleUpdateInput{
		RoleID:      id,
		Name:        req.Name,
		Permissions: req.Permissions,
		ActorUserID: actorUserIDFromContext(c),
	})
	if updateErr != nil {
		apperror.Write(c, updateErr)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *AdminHandler) UpdateRolePermissions(c *gin.Context) {
	id, err := parseRoleIDParam(c.Param("id"))
	if err != nil {
		apperror.Write(c, err)
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		apperror.Write(c, apperror.ValidationWithDetails("validation failed", map[string]any{"body": []string{"invalid JSON payload"}}))
		return
	}

	updated, updateErr := h.service.UpdateRole(c.Request.Context(), RoleUpdateInput{
		RoleID:      id,
		Permissions: &req.Permissions,
		ActorUserID: actorUserIDFromContext(c),
	})
	if updateErr != nil {
		apperror.Write(c, updateErr)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *AdminHandler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "25"))
	sortField, sortOrder := parseSortQuery(c.Query("sort"))
	filterField, filterValue := parseFilterQuery(c.Query("filter"))

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		query = strings.TrimSpace(filterForPermissions(filterField, filterValue))
	}

	var moduleScope *string
	moduleValue := strings.TrimSpace(c.Query("moduleScope"))
	if moduleValue != "" {
		moduleScope = &moduleValue
	}

	result, err := h.service.ListPermissions(c.Request.Context(), PermissionListQuery{
		Page:        page,
		PageSize:    pageSize,
		SortField:   sortField,
		SortOrder:   sortOrder,
		Query:       query,
		ModuleScope: moduleScope,
	})
	if err != nil {
		apperror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      result.Items,
		"totalCount": result.Total,
		"page":       result.Page,
		"pageSize":   result.PageSize,
	})
}

func parseRoleIDParam(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"invalid role id"}})
	}
	return id, nil
}

func parseBoolQuery(raw string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	return trimmed == "1" || trimmed == "true" || trimmed == "yes"
}

func parseSortQuery(raw string) (field string, order string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), "asc"
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func parseFilterQuery(raw string) (field string, value string) {
	filter := strings.TrimSpace(raw)
	if filter == "" {
		return "", ""
	}
	parts := strings.SplitN(filter, ":", 2)
	if len(parts) == 1 {
		return "", strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func filterForRoles(field, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if field == "" || field == "name" {
		return value
	}
	return ""
}

func filterForPermissions(field, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if field == "" || field == "name" || field == "moduleScope" {
		return value
	}
	return ""
}

func actorUserIDFromContext(c *gin.Context) *int64 {
	value, ok := c.Get("actor_user_id")
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case int64:
		return &typed
	case int:
		v := int64(typed)
		return &v
	default:
		return nil
	}
}
