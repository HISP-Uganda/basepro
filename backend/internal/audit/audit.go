package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Event struct {
	Timestamp   time.Time
	ActorUserID *int64
	Action      string
	EntityType  string
	EntityID    *string
	Metadata    map[string]any
}

type Repository interface {
	Insert(ctx context.Context, event Event) error
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Insert(ctx context.Context, event Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO audit_logs (timestamp, actor_user_id, action, entity_type, entity_id, metadata_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.Timestamp, event.ActorUserID, event.Action, event.EntityType, event.EntityID, metadata, event.Timestamp)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}
	return s.repo.Insert(ctx, event)
}
