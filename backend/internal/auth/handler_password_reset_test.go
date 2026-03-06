package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestForgotPasswordEndpointReturnsAcceptedForUnknownAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepo(&User{ID: 1, Username: "admin", IsActive: true})
	handler := NewHandler(newTestService(repo, &fakeAuditRepo{}))

	r := gin.New()
	r.POST("/api/v1/auth/forgot-password", handler.ForgotPassword)

	payload := map[string]any{"identifier": "missing-user"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestForgotPasswordEndpointDoesNotLeakExistingAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepo(&User{ID: 2, Username: "alice", IsActive: true})
	handler := NewHandler(newTestService(repo, &fakeAuditRepo{}))

	r := gin.New()
	r.POST("/api/v1/auth/forgot-password", handler.ForgotPassword)

	payload := map[string]any{"identifier": "alice"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", w.Code, w.Body.String())
	}
}
