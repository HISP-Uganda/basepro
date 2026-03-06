package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeListRepo struct {
	lastFilter ListFilter
}

func (f *fakeListRepo) Insert(context.Context, Event) error {
	return nil
}

func (f *fakeListRepo) List(_ context.Context, filter ListFilter) (ListResult, error) {
	f.lastFilter = filter
	return ListResult{Items: []Record{}, Total: 0, Page: filter.Page, PageSize: filter.PageSize}, nil
}

func TestAuditHandlerListSupportsQSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeListRepo{}
	handler := NewHandler(NewService(repo))

	r := gin.New()
	r.GET("/audit", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/audit?page=3&pageSize=10&q=auth.login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.Page != 3 || repo.lastFilter.PageSize != 10 {
		t.Fatalf("expected page/pageSize 3/10, got %d/%d", repo.lastFilter.Page, repo.lastFilter.PageSize)
	}
	if repo.lastFilter.Action != "auth.login" {
		t.Fatalf("expected q to map to action filter, got %q", repo.lastFilter.Action)
	}
}

func TestAuditHandlerListRejectsInvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(NewService(&fakeListRepo{}))

	r := gin.New()
	r.GET("/audit", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/audit?page=bad", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errorBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope")
	}
	if errorBody["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %v", errorBody["code"])
	}
}
