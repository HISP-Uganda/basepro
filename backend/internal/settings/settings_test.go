package settings

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"basepro/backend/internal/audit"
)

type fakeRepo struct {
	items map[string]json.RawMessage
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: map[string]json.RawMessage{}}
}

func (f *fakeRepo) key(category, key string) string {
	return category + "::" + key
}

func (f *fakeRepo) Get(_ context.Context, category, key string) (json.RawMessage, error) {
	value, ok := f.items[f.key(category, key)]
	if !ok {
		return nil, ErrNotFound
	}
	return append(json.RawMessage{}, value...), nil
}

func (f *fakeRepo) Upsert(_ context.Context, category, key string, value json.RawMessage, _ *int64, _ time.Time) error {
	f.items[f.key(category, key)] = append(json.RawMessage{}, value...)
	return nil
}

type fakeAuditRepo struct {
	events []audit.Event
}

func (f *fakeAuditRepo) Insert(_ context.Context, event audit.Event) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeAuditRepo) List(context.Context, audit.ListFilter) (audit.ListResult, error) {
	return audit.ListResult{}, nil
}

func TestGetLoginBrandingDefaultsWhenMissing(t *testing.T) {
	service := NewService(newFakeRepo(), nil)

	got, err := service.GetLoginBranding(context.Background())
	if err != nil {
		t.Fatalf("get branding: %v", err)
	}
	if got.ApplicationDisplayName != "BasePro" {
		t.Fatalf("expected BasePro default, got %q", got.ApplicationDisplayName)
	}
	if got.ImageConfigured {
		t.Fatal("expected imageConfigured false")
	}
}

func TestUpdateAndGetLoginBranding(t *testing.T) {
	repo := newFakeRepo()
	auditRepo := &fakeAuditRepo{}
	service := NewService(repo, audit.NewService(auditRepo))
	actor := int64(42)

	url := "https://cdn.example.com/brand.png"
	asset := "assets/login/custom.png"
	updated, err := service.UpdateLoginBranding(context.Background(), LoginBrandingUpdateInput{
		ApplicationDisplayName: "BasePro Ops",
		LoginImageURL:          &url,
		LoginImageAssetPath:    &asset,
	}, &actor)
	if err != nil {
		t.Fatalf("update branding: %v", err)
	}
	if updated.ApplicationDisplayName != "BasePro Ops" {
		t.Fatalf("unexpected display name: %q", updated.ApplicationDisplayName)
	}
	if !updated.ImageConfigured {
		t.Fatal("expected imageConfigured true")
	}

	got, err := service.GetLoginBranding(context.Background())
	if err != nil {
		t.Fatalf("get branding: %v", err)
	}
	if got.ApplicationDisplayName != "BasePro Ops" {
		t.Fatalf("expected persisted display name, got %q", got.ApplicationDisplayName)
	}
	if got.LoginImageURL == nil || *got.LoginImageURL != url {
		t.Fatalf("expected persisted loginImageUrl %q", url)
	}

	if len(auditRepo.events) != 1 || auditRepo.events[0].Action != "settings.login_branding.update" {
		t.Fatalf("expected settings.login_branding.update audit event, got %+v", auditRepo.events)
	}
}

func TestUpdateLoginBrandingRejectsBadURL(t *testing.T) {
	service := NewService(newFakeRepo(), nil)
	bad := "javascript:alert(1)"
	_, err := service.UpdateLoginBranding(context.Background(), LoginBrandingUpdateInput{
		ApplicationDisplayName: "BasePro",
		LoginImageURL:          &bad,
	}, nil)
	if err == nil {
		t.Fatal("expected validation error for bad URL")
	}
}
