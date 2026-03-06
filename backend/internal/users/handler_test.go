package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"basepro/backend/internal/audit"
	"basepro/backend/internal/auth"
	"basepro/backend/internal/rbac"
	"github.com/gin-gonic/gin"
)

func newTestHandler() *Handler {
	handler, _ := newTestHandlerWithRepo()
	return handler
}

func newTestHandlerWithRepo() (*Handler, *fakeRepo) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	service := NewService(repo, rbacService, audit.NewService(&fakeAuditRepo{}), 4)
	return NewHandler(service), repo
}

func withPrincipal(c *gin.Context) {
	c.Set(auth.PrincipalContextKey, auth.Principal{Type: "user", UserID: 1, Username: "admin"})
}

func TestCreateUserEndpointReturnsMetadataWithoutPasswordHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler()

	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		withPrincipal(c)
		handler.Create(c)
	})

	payload := map[string]any{
		"username":    "meta-user",
		"password":    "TempPass123!",
		"email":       "meta@example.com",
		"firstName":   "Meta",
		"lastName":    "User",
		"phoneNumber": "+15551234567",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", w.Code, w.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, exists := response["passwordHash"]; exists {
		t.Fatalf("passwordHash must not be returned")
	}
	if response["email"] != "meta@example.com" {
		t.Fatalf("expected email in response")
	}
}

func TestCreateUserValidationErrorUsesStandardizedShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler()

	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		withPrincipal(c)
		handler.Create(c)
	})

	payload := map[string]any{
		"username": "bad",
		"password": "TempPass123!",
		"email":    "bad-email",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", w.Code, w.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errorBody, ok := response["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope")
	}
	if errorBody["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %v", errorBody["code"])
	}
	details, ok := errorBody["details"].(map[string]any)
	if !ok {
		t.Fatalf("expected details object")
	}
	if _, ok := details["email"]; !ok {
		t.Fatalf("expected details.email")
	}
}

func TestListUsersSupportsQSearchFallbackAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newTestHandlerWithRepo()
	repo.users[1] = UserRecord{ID: 1, Username: "alice", Language: "English", IsActive: true}
	repo.users[2] = UserRecord{ID: 2, Username: "bob", Language: "English", IsActive: true}

	r := gin.New()
	r.GET("/users", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/users?page=2&pageSize=10&q=ali", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	if repo.lastListQuery.Page != 2 || repo.lastListQuery.PageSize != 10 {
		t.Fatalf("expected query page/pageSize 2/10, got %d/%d", repo.lastListQuery.Page, repo.lastListQuery.PageSize)
	}
	if repo.lastListQuery.Filter != "ali" {
		t.Fatalf("expected q to map into filter, got %q", repo.lastListQuery.Filter)
	}
}

func TestListUsersRejectsInvalidSortQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHandler()

	r := gin.New()
	r.GET("/users", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/users?sort=username:sideways", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", w.Code, w.Body.String())
	}
}
