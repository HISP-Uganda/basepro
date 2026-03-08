package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"basepro/backend/internal/auth"
	"basepro/backend/internal/bootstrap"
	"basepro/backend/internal/moduleenablement"
	"basepro/backend/internal/rbac"
	"basepro/backend/internal/settings"
)

func TestBootstrapEndpointReturnsTypedPayload(t *testing.T) {
	moduleService := moduleenablement.NewService(&fakeSettingsRepo{}, nil)
	settingsService := settings.NewService(&fakeSettingsRepo{}, nil)
	bootstrapService := bootstrap.NewService(
		bootstrap.AppInfo{Version: "1.2.3", Commit: "abc123", BuildDate: "2026-03-08"},
		func() bootstrap.RuntimeInfo {
			return bootstrap.RuntimeInfo{
				Environment:       "test",
				APITokenEnabled:   true,
				APIHeaderName:     "X-API-Token",
				APITokenAllowAuth: false,
			}
		},
		settingsService,
		moduleService,
		func() map[string]bool {
			return map[string]bool{"modules.settings.enabled": false}
		},
		nil,
	)

	router := newRouter(AppDeps{
		BootstrapHandler: bootstrap.NewHandler(bootstrapService),
		JWTManager:       auth.NewJWTManager("jwt-secret", time.Minute),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["version"] != float64(1) {
		t.Fatalf("expected version=1, got %v", body["version"])
	}

	app, _ := body["app"].(map[string]any)
	if app["version"] != "1.2.3" {
		t.Fatalf("expected app.version 1.2.3, got %v", app["version"])
	}

	capabilities, _ := body["capabilities"].(map[string]any)
	settingsCaps, _ := capabilities["settings"].(map[string]any)
	if settingsCaps["authorization"] != "admin_or_settings.write" {
		t.Fatalf("expected settings authorization contract, got %v", settingsCaps["authorization"])
	}

	modules, _ := body["modules"].([]any)
	if len(modules) == 0 {
		t.Fatal("expected non-empty modules in bootstrap response")
	}
}

func TestBootstrapEndpointIncludesAuthenticatedUserSummary(t *testing.T) {
	jwt := auth.NewJWTManager("jwt-secret", time.Minute)
	token, _, _ := jwt.GenerateAccessToken(88, "writer", time.Now().UTC())

	rbacService := rbac.NewService(&fakeRBACRepo{
		rolesByUser: map[int64][]rbac.Role{
			88: {{ID: 1, Name: "Manager"}},
		},
		permsByUser: map[int64][]rbac.Permission{
			88: {
				{ID: 1, Name: "settings.read"},
				{ID: 2, Name: "settings.write"},
			},
		},
	})

	moduleService := moduleenablement.NewService(&fakeSettingsRepo{}, nil)
	settingsService := settings.NewService(&fakeSettingsRepo{}, nil)
	bootstrapService := bootstrap.NewService(
		bootstrap.AppInfo{Version: "1.2.3", Commit: "abc123", BuildDate: "2026-03-08"},
		func() bootstrap.RuntimeInfo { return bootstrap.RuntimeInfo{} },
		settingsService,
		moduleService,
		nil,
		rbacService,
	)

	router := newRouter(AppDeps{
		JWTManager:       jwt,
		RBACService:      rbacService,
		BootstrapHandler: bootstrap.NewHandler(bootstrapService),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	principal, ok := body["principal"].(map[string]any)
	if !ok {
		t.Fatalf("expected principal in bootstrap payload: %v", body["principal"])
	}
	if principal["username"] != "writer" {
		t.Fatalf("expected principal username writer, got %v", principal["username"])
	}

	capabilities, _ := body["capabilities"].(map[string]any)
	settingsCaps, _ := capabilities["settings"].(map[string]any)
	if settingsCaps["canWrite"] != true {
		t.Fatalf("expected settings canWrite=true for settings.write principal, got %v", settingsCaps["canWrite"])
	}
}
