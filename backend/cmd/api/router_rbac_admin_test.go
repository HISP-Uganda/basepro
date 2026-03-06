package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"basepro/backend/internal/audit"
	"basepro/backend/internal/auth"
	"basepro/backend/internal/rbac"
)

type fakeRBACRepo struct {
	rolesByUser map[int64][]rbac.Role
	permsByUser map[int64][]rbac.Permission
}

func (f *fakeRBACRepo) GetUserRoles(_ context.Context, userID int64) ([]rbac.Role, error) {
	return append([]rbac.Role{}, f.rolesByUser[userID]...), nil
}

func (f *fakeRBACRepo) GetUserPermissions(_ context.Context, userID int64) ([]rbac.Permission, error) {
	return append([]rbac.Permission{}, f.permsByUser[userID]...), nil
}

func (f *fakeRBACRepo) EnsureRole(context.Context, string) (rbac.Role, error) {
	return rbac.Role{}, nil
}
func (f *fakeRBACRepo) EnsurePermission(context.Context, string, *string) (rbac.Permission, error) {
	return rbac.Permission{}, nil
}
func (f *fakeRBACRepo) EnsureRolePermission(context.Context, int64, int64) error { return nil }
func (f *fakeRBACRepo) EnsureUserRole(context.Context, int64, int64) error       { return nil }
func (f *fakeRBACRepo) GetRoleByName(context.Context, string) (rbac.Role, error) {
	return rbac.Role{}, nil
}
func (f *fakeRBACRepo) ReplaceUserRoles(context.Context, int64, []int64) error { return nil }

type fakeAdminRepo struct{}

func (f *fakeAdminRepo) ListRoles(_ context.Context, query rbac.RoleListQuery) (rbac.RoleListResult, error) {
	return rbac.RoleListResult{
		Items:    []rbac.RoleSummary{},
		Total:    0,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
func (f *fakeAdminRepo) CreateRole(context.Context, string) (rbac.RoleRecord, error) {
	now := time.Now().UTC()
	return rbac.RoleRecord{ID: 1, Name: "New", CreatedAt: now, UpdatedAt: now}, nil
}
func (f *fakeAdminRepo) UpdateRoleName(context.Context, int64, string) (rbac.RoleRecord, error) {
	now := time.Now().UTC()
	return rbac.RoleRecord{ID: 1, Name: "Updated", CreatedAt: now, UpdatedAt: now}, nil
}
func (f *fakeAdminRepo) GetRoleByID(context.Context, int64) (rbac.RoleRecord, error) {
	now := time.Now().UTC()
	return rbac.RoleRecord{ID: 1, Name: "Admin", CreatedAt: now, UpdatedAt: now}, nil
}
func (f *fakeAdminRepo) ListRolePermissions(context.Context, int64) ([]rbac.PermissionRecord, error) {
	return []rbac.PermissionRecord{}, nil
}
func (f *fakeAdminRepo) ListRoleUsers(context.Context, int64) ([]rbac.RoleUserRecord, error) {
	return []rbac.RoleUserRecord{}, nil
}
func (f *fakeAdminRepo) ListPermissions(_ context.Context, query rbac.PermissionListQuery) (rbac.PermissionListResult, error) {
	return rbac.PermissionListResult{Items: []rbac.PermissionRecord{}, Total: 0, Page: query.Page, PageSize: query.PageSize}, nil
}
func (f *fakeAdminRepo) GetPermissionsByNames(context.Context, []string) ([]rbac.PermissionRecord, error) {
	return []rbac.PermissionRecord{}, nil
}
func (f *fakeAdminRepo) ReplaceRolePermissions(context.Context, int64, []int64) error { return nil }

type fakeAuditRepo struct{}

func (f *fakeAuditRepo) Insert(context.Context, audit.Event) error { return nil }
func (f *fakeAuditRepo) List(context.Context, audit.ListFilter) (audit.ListResult, error) {
	return audit.ListResult{}, nil
}

func TestRBACAdminRoutesRequireAuthentication(t *testing.T) {
	jwt := auth.NewJWTManager("jwt-secret", time.Minute)
	rbacService := rbac.NewService(&fakeRBACRepo{})
	adminService := rbac.NewAdminService(&fakeAdminRepo{}, audit.NewService(&fakeAuditRepo{}))

	router := newRouter(AppDeps{
		JWTManager:       jwt,
		RBACService:      rbacService,
		RBACAdminHandler: rbac.NewAdminHandler(adminService),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var body map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"]["code"] != "AUTH_UNAUTHORIZED" {
		t.Fatalf("expected AUTH_UNAUTHORIZED, got %v", body["error"]["code"])
	}
}

func TestRBACAdminRoutesEnforcePermissionForWrite(t *testing.T) {
	jwt := auth.NewJWTManager("jwt-secret", time.Minute)
	token, _, _ := jwt.GenerateAccessToken(10, "reader", time.Now().UTC())

	rbacService := rbac.NewService(&fakeRBACRepo{
		rolesByUser: map[int64][]rbac.Role{10: []rbac.Role{{ID: 1, Name: "Viewer"}}},
		permsByUser: map[int64][]rbac.Permission{10: []rbac.Permission{{ID: 1, Name: "users.read"}}},
	})
	adminService := rbac.NewAdminService(&fakeAdminRepo{}, audit.NewService(&fakeAuditRepo{}))

	router := newRouter(AppDeps{
		JWTManager:       jwt,
		RBACService:      rbacService,
		RBACAdminHandler: rbac.NewAdminHandler(adminService),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}
