package audit

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/listquery"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

	action := strings.TrimSpace(c.Query("action"))
	if action == "" {
		action = listquery.ResolveSearch(c.Query("q"), filterField, filterValue, map[string]struct{}{"action": {}})
	}

	var actorUserID *int64
	actorRaw := strings.TrimSpace(c.Query("actor_user_id"))
	if actorRaw == "" {
		actorRaw = strings.TrimSpace(c.Query("actorUserId"))
	}
	if actorRaw != "" {
		parsedActor, err := strconv.ParseInt(actorRaw, 10, 64)
		if err != nil {
			apperror.Write(c, apperror.Validation("actor_user_id must be an integer"))
			return
		}
		actorUserID = &parsedActor
	}

	dateFrom, err := parseDateQuery(c.Query("date_from"))
	if err != nil {
		apperror.Write(c, apperror.Validation("date_from must be RFC3339 or YYYY-MM-DD"))
		return
	}
	if dateFrom == nil {
		dateFrom, err = parseDateQuery(c.Query("dateFrom"))
		if err != nil {
			apperror.Write(c, apperror.Validation("dateFrom must be RFC3339 or YYYY-MM-DD"))
			return
		}
	}

	dateTo, err := parseDateQuery(c.Query("date_to"))
	if err != nil {
		apperror.Write(c, apperror.Validation("date_to must be RFC3339 or YYYY-MM-DD"))
		return
	}
	if dateTo == nil {
		dateTo, err = parseDateQuery(c.Query("dateTo"))
		if err != nil {
			apperror.Write(c, apperror.Validation("dateTo must be RFC3339 or YYYY-MM-DD"))
			return
		}
	}

	result, err := h.service.List(c.Request.Context(), ListFilter{
		Page:        page,
		PageSize:    pageSize,
		SortField:   sortField,
		SortOrder:   sortOrder,
		Action:      action,
		ActorUserID: actorUserID,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
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

func parseDateQuery(raw string) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		t := parsed.UTC()
		return &t, nil
	}
	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		t := parsed.UTC()
		return &t, nil
	}
	return nil, errors.New("invalid date format")
}
