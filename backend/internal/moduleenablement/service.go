package moduleenablement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/audit"
	"basepro/backend/internal/settings"
)

const (
	settingsCategory = "module_enablement"
	settingsKey      = "runtime_overrides"
)

type SettingsRepository interface {
	Get(ctx context.Context, category, key string) (json.RawMessage, error)
	Upsert(ctx context.Context, category, key string, value json.RawMessage, updatedByUserID *int64, now time.Time) error
}

type RuntimeModuleOverride struct {
	ModuleID string `json:"moduleId"`
	Enabled  bool   `json:"enabled"`
}

type runtimeOverridesStored struct {
	Flags map[string]bool `json:"flags"`
}

type Service struct {
	repo         SettingsRepository
	auditService *audit.Service
	now          func() time.Time
	runtime      atomic.Value
}

func NewService(repo SettingsRepository, auditService *audit.Service) *Service {
	svc := &Service{
		repo:         repo,
		auditService: auditService,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	svc.runtime.Store(map[string]bool{})
	return svc
}

func (s *Service) LoadRuntimeOverrides(ctx context.Context) error {
	if s.repo == nil {
		s.runtime.Store(map[string]bool{})
		return nil
	}
	raw, err := s.repo.Get(ctx, settingsCategory, settingsKey)
	if err != nil {
		if errors.Is(err, settings.ErrNotFound) {
			s.runtime.Store(map[string]bool{})
			return nil
		}
		return err
	}

	var stored runtimeOverridesStored
	if unmarshalErr := json.Unmarshal(raw, &stored); unmarshalErr != nil {
		return fmt.Errorf("decode module runtime overrides: %w", unmarshalErr)
	}
	if err := validateRuntimeOverrides(stored.Flags); err != nil {
		return err
	}
	s.runtime.Store(cloneMap(stored.Flags))
	return nil
}

func (s *Service) ListEffective(configOverrides map[string]bool) []EffectiveModule {
	return ResolveEffective(configOverrides, s.getRuntimeOverrides())
}

func (s *Service) EffectiveOverrideMap(configOverrides map[string]bool) map[string]bool {
	merged := make(map[string]bool, len(configOverrides)+len(s.getRuntimeOverrides()))
	for key, value := range configOverrides {
		merged[key] = value
	}
	for key, value := range s.getRuntimeOverrides() {
		merged[key] = value
	}
	return merged
}

func (s *Service) UpdateRuntimeOverrides(
	ctx context.Context,
	updates []RuntimeModuleOverride,
	configOverrides map[string]bool,
	actorUserID *int64,
) ([]EffectiveModule, error) {
	if len(updates) == 0 {
		return nil, apperror.ValidationWithDetails("validation failed", map[string]any{
			"modules": []string{"at least one module override is required"},
		})
	}

	details := map[string][]string{}
	nextRuntime := s.getRuntimeOverrides()
	serialized := make([]map[string]any, 0, len(updates))
	seenModules := map[string]struct{}{}

	for _, update := range updates {
		definition, ok := EditableDefinitionByModuleID(update.ModuleID)
		if !ok {
			details[update.ModuleID] = []string{"module is not runtime-manageable"}
			continue
		}
		if _, duplicate := seenModules[update.ModuleID]; duplicate {
			details[update.ModuleID] = []string{"duplicate module update"}
			continue
		}
		seenModules[update.ModuleID] = struct{}{}

		nextRuntime[definition.FlagKey] = update.Enabled
		serialized = append(serialized, map[string]any{
			"moduleId": update.ModuleID,
			"enabled":  update.Enabled,
		})
	}

	if len(details) > 0 {
		return nil, apperror.ValidationWithDetails("validation failed", map[string]any{
			"modules": details,
		})
	}

	stored := runtimeOverridesStored{Flags: nextRuntime}
	payload, err := json.Marshal(stored)
	if err != nil {
		return nil, err
	}
	if s.repo == nil {
		return nil, fmt.Errorf("module enablement repository is not configured")
	}
	now := s.now()
	if err := s.repo.Upsert(ctx, settingsCategory, settingsKey, payload, actorUserID, now); err != nil {
		return nil, err
	}

	s.runtime.Store(cloneMap(nextRuntime))
	s.logAudit(ctx, audit.Event{
		Action:      "settings.module_enablement.update",
		ActorUserID: actorUserID,
		EntityType:  "settings",
		EntityID:    strPtr("module_enablement.runtime_overrides"),
		Metadata: map[string]any{
			"updates": serialized,
		},
	})

	return ResolveEffective(configOverrides, nextRuntime), nil
}

func (s *Service) getRuntimeOverrides() map[string]bool {
	current, ok := s.runtime.Load().(map[string]bool)
	if !ok {
		return map[string]bool{}
	}
	return cloneMap(current)
}

func validateRuntimeOverrides(overrides map[string]bool) error {
	if len(overrides) == 0 {
		return nil
	}
	for key := range overrides {
		allowed := false
		for _, definition := range Definitions() {
			if definition.FlagKey == key && isRuntimeEditable(definition) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("module runtime overrides contains non-editable key %q", key)
		}
	}
	return nil
}

func cloneMap(input map[string]bool) map[string]bool {
	if len(input) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func (s *Service) logAudit(ctx context.Context, event audit.Event) {
	if s.auditService == nil {
		return
	}
	_ = s.auditService.Log(ctx, event)
}

func strPtr(v string) *string {
	return &v
}
