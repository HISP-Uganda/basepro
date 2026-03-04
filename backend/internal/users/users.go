package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/audit"
	"basepro/backend/internal/auth"
	"basepro/backend/internal/rbac"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type UserRecord struct {
	ID             int64      `db:"id" json:"id"`
	Username       string     `db:"username" json:"username"`
	Email          *string    `db:"email" json:"email,omitempty"`
	Language       string     `db:"language" json:"language"`
	FirstName      *string    `db:"first_name" json:"firstName,omitempty"`
	LastName       *string    `db:"last_name" json:"lastName,omitempty"`
	DisplayName    *string    `db:"display_name" json:"displayName,omitempty"`
	PhoneNumber    *string    `db:"phone_number" json:"phoneNumber,omitempty"`
	WhatsappNumber *string    `db:"whatsapp_number" json:"whatsappNumber,omitempty"`
	TelegramHandle *string    `db:"telegram_handle" json:"telegramHandle,omitempty"`
	IsActive       bool       `db:"is_active" json:"isActive"`
	LastLoginAt    *time.Time `db:"last_login_at" json:"lastLoginAt,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
	Roles          []string   `json:"roles"`
}

type ListQuery struct {
	Page      int
	PageSize  int
	SortField string
	SortOrder string
	Filter    string
}

type ListResult struct {
	Items    []UserRecord
	Total    int
	Page     int
	PageSize int
}

type CreateUserParams struct {
	Username       string
	PasswordHash   string
	Email          *string
	Language       string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	IsActive       bool
}

type UpdateUserParams struct {
	UserID         int64
	Username       *string
	PasswordHash   *string
	Email          *string
	Language       *string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	IsActive       *bool
}

type Repository interface {
	ListUsers(ctx context.Context, query ListQuery) (ListResult, error)
	GetUserByID(ctx context.Context, userID int64) (UserRecord, error)
	CreateUser(ctx context.Context, params CreateUserParams) (UserRecord, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (UserRecord, error)
	SetPassword(ctx context.Context, userID int64, passwordHash string) error
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func normalizeListQuery(query ListQuery) ListQuery {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 25
	}
	sortField := strings.TrimSpace(query.SortField)
	switch sortField {
	case "id", "username", "email", "display_name", "created_at", "updated_at", "last_login_at", "is_active":
	default:
		sortField = "id"
	}
	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	return ListQuery{
		Page:      page,
		PageSize:  pageSize,
		SortField: sortField,
		SortOrder: sortOrder,
		Filter:    strings.TrimSpace(query.Filter),
	}
}

func (r *SQLRepository) ListUsers(ctx context.Context, query ListQuery) (ListResult, error) {
	q := normalizeListQuery(query)
	offset := (q.Page - 1) * q.PageSize
	filterValue := "%"
	hasFilter := q.Filter != ""
	if hasFilter {
		filterValue = "%" + q.Filter + "%"
	}

	total := 0
	countArgs := []any{}
	countQuery := `SELECT COUNT(*) FROM users`
	if hasFilter {
		countQuery += ` WHERE username ILIKE $1 OR COALESCE(email, '') ILIKE $1 OR COALESCE(display_name, '') ILIKE $1`
		countArgs = append(countArgs, filterValue)
	}
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return ListResult{}, fmt.Errorf("count users: %w", err)
	}

	items := []UserRecord{}
	selectArgs := []any{q.PageSize, offset}
	selectQuery := `
		SELECT id, username, email, language, first_name, last_name, display_name,
		       phone_number, whatsapp_number, telegram_handle, is_active, last_login_at,
		       created_at, updated_at
		FROM users
	`
	if hasFilter {
		selectQuery += ` WHERE username ILIKE $3 OR COALESCE(email, '') ILIKE $3 OR COALESCE(display_name, '') ILIKE $3`
		selectArgs = append(selectArgs, filterValue)
	}
	selectQuery += fmt.Sprintf(" ORDER BY %s %s LIMIT $1 OFFSET $2", q.SortField, strings.ToUpper(q.SortOrder))

	if err := r.db.SelectContext(ctx, &items, selectQuery, selectArgs...); err != nil {
		return ListResult{}, fmt.Errorf("list users: %w", err)
	}
	for i := range items {
		items[i].Roles = []string{}
	}

	return ListResult{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (r *SQLRepository) GetUserByID(ctx context.Context, userID int64) (UserRecord, error) {
	var record UserRecord
	err := r.db.GetContext(ctx, &record, `
		SELECT id, username, email, language, first_name, last_name, display_name,
		       phone_number, whatsapp_number, telegram_handle, is_active, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, sql.ErrNoRows
		}
		return UserRecord{}, fmt.Errorf("get user by id: %w", err)
	}
	record.Roles = []string{}
	return record, nil
}

func (r *SQLRepository) CreateUser(ctx context.Context, params CreateUserParams) (UserRecord, error) {
	var record UserRecord
	err := r.db.GetContext(ctx, &record, `
		INSERT INTO users (
			username, password_hash, email, language, first_name, last_name, display_name,
			phone_number, whatsapp_number, telegram_handle, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, username, email, language, first_name, last_name, display_name,
		          phone_number, whatsapp_number, telegram_handle, is_active, last_login_at,
		          created_at, updated_at
	`,
		params.Username,
		params.PasswordHash,
		params.Email,
		params.Language,
		params.FirstName,
		params.LastName,
		params.DisplayName,
		params.PhoneNumber,
		params.WhatsappNumber,
		params.TelegramHandle,
		params.IsActive,
	)
	if err != nil {
		return UserRecord{}, fmt.Errorf("create user: %w", err)
	}
	record.Roles = []string{}
	return record, nil
}

func (r *SQLRepository) UpdateUser(ctx context.Context, params UpdateUserParams) (UserRecord, error) {
	sets := []string{}
	args := []any{}

	if params.Username != nil {
		args = append(args, *params.Username)
		sets = append(sets, fmt.Sprintf("username = $%d", len(args)))
	}
	if params.PasswordHash != nil {
		args = append(args, *params.PasswordHash)
		sets = append(sets, fmt.Sprintf("password_hash = $%d", len(args)))
	}
	if params.Email != nil {
		args = append(args, params.Email)
		sets = append(sets, fmt.Sprintf("email = $%d", len(args)))
	}
	if params.Language != nil {
		args = append(args, *params.Language)
		sets = append(sets, fmt.Sprintf("language = $%d", len(args)))
	}
	if params.FirstName != nil {
		args = append(args, params.FirstName)
		sets = append(sets, fmt.Sprintf("first_name = $%d", len(args)))
	}
	if params.LastName != nil {
		args = append(args, params.LastName)
		sets = append(sets, fmt.Sprintf("last_name = $%d", len(args)))
	}
	if params.DisplayName != nil {
		args = append(args, params.DisplayName)
		sets = append(sets, fmt.Sprintf("display_name = $%d", len(args)))
	}
	if params.PhoneNumber != nil {
		args = append(args, params.PhoneNumber)
		sets = append(sets, fmt.Sprintf("phone_number = $%d", len(args)))
	}
	if params.WhatsappNumber != nil {
		args = append(args, params.WhatsappNumber)
		sets = append(sets, fmt.Sprintf("whatsapp_number = $%d", len(args)))
	}
	if params.TelegramHandle != nil {
		args = append(args, params.TelegramHandle)
		sets = append(sets, fmt.Sprintf("telegram_handle = $%d", len(args)))
	}
	if params.IsActive != nil {
		args = append(args, *params.IsActive)
		sets = append(sets, fmt.Sprintf("is_active = $%d", len(args)))
	}

	if len(sets) == 0 {
		return r.GetUserByID(ctx, params.UserID)
	}

	args = append(args, params.UserID)
	query := fmt.Sprintf(`
		UPDATE users
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, username, email, language, first_name, last_name, display_name,
		          phone_number, whatsapp_number, telegram_handle, is_active, last_login_at,
		          created_at, updated_at
	`, strings.Join(sets, ", "), len(args))

	var record UserRecord
	if err := r.db.GetContext(ctx, &record, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, sql.ErrNoRows
		}
		return UserRecord{}, fmt.Errorf("update user: %w", err)
	}
	record.Roles = []string{}
	return record, nil
}

func (r *SQLRepository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type Service struct {
	repo             Repository
	rbacService      *rbac.Service
	auditService     *audit.Service
	passwordHashCost int
}

func NewService(repo Repository, rbacService *rbac.Service, auditService *audit.Service, passwordHashCost int) *Service {
	return &Service{repo: repo, rbacService: rbacService, auditService: auditService, passwordHashCost: passwordHashCost}
}

func (s *Service) ListUsers(ctx context.Context, query ListQuery) (ListResult, error) {
	users, err := s.repo.ListUsers(ctx, query)
	if err != nil {
		return ListResult{}, err
	}
	for i := range users.Items {
		roles, roleErr := s.rbacService.RoleNamesForUser(ctx, users.Items[i].ID)
		if roleErr != nil {
			return ListResult{}, roleErr
		}
		users.Items[i].Roles = roles
	}
	return users, nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (UserRecord, error) {
	record, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"user not found"}})
		}
		return UserRecord{}, err
	}
	roles, roleErr := s.rbacService.RoleNamesForUser(ctx, record.ID)
	if roleErr != nil {
		return UserRecord{}, roleErr
	}
	record.Roles = roles
	return record, nil
}

type CreateInput struct {
	Username       string
	Password       string
	Email          *string
	Language       *string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	IsActive       bool
	Roles          []string
	ActorID        *int64
}

func (s *Service) CreateUser(ctx context.Context, in CreateInput) (UserRecord, error) {
	normalized, validationDetails := normalizeCreateInput(in)
	if len(validationDetails) > 0 {
		return UserRecord{}, apperror.ValidationWithDetails("validation failed", validationDetails)
	}

	hash, err := auth.HashPassword(normalized.Password, s.passwordHashCost)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"roles": []string{"one or more roles are invalid"}})
		}
		return UserRecord{}, err
	}

	created, err := s.repo.CreateUser(ctx, CreateUserParams{
		Username:       normalized.Username,
		PasswordHash:   hash,
		Email:          normalized.Email,
		Language:       normalized.Language,
		FirstName:      normalized.FirstName,
		LastName:       normalized.LastName,
		DisplayName:    normalized.DisplayName,
		PhoneNumber:    normalized.PhoneNumber,
		WhatsappNumber: normalized.WhatsappNumber,
		TelegramHandle: normalized.TelegramHandle,
		IsActive:       normalized.IsActive,
	})
	if err != nil {
		if mapped := mapConstraintError(err); mapped != nil {
			return UserRecord{}, mapped
		}
		return UserRecord{}, err
	}

	for _, role := range normalized.Roles {
		if assignErr := s.rbacService.AssignRoleToUser(ctx, created.ID, role); assignErr != nil {
			if errors.Is(assignErr, sql.ErrNoRows) {
				return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"roles": []string{"one or more roles are invalid"}})
			}
			return UserRecord{}, assignErr
		}
	}
	r, err := s.rbacService.RoleNamesForUser(ctx, created.ID)
	if err != nil {
		return UserRecord{}, err
	}
	created.Roles = r

	s.logAudit(ctx, audit.Event{Action: "users.create", ActorUserID: in.ActorID, EntityType: "user", EntityID: strPtr(created.Username), Metadata: map[string]any{"user_id": created.ID, "roles": created.Roles}})
	return created, nil
}

type UpdateInput struct {
	UserID         int64
	Username       *string
	Password       *string
	Email          *string
	Language       *string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	Roles          *[]string
	IsActive       *bool
	ActorID        *int64
}

func (s *Service) UpdateUser(ctx context.Context, in UpdateInput) (UserRecord, error) {
	existing, err := s.repo.GetUserByID(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"user not found"}})
		}
		return UserRecord{}, err
	}

	normalized, validationDetails := normalizeUpdateInput(in, existing)
	if len(validationDetails) > 0 {
		return UserRecord{}, apperror.ValidationWithDetails("validation failed", validationDetails)
	}

	metadata := map[string]any{}
	params := UpdateUserParams{UserID: in.UserID}
	if normalized.Username != nil {
		params.Username = normalized.Username
		metadata["username"] = *normalized.Username
	}
	if normalized.Email != nil {
		params.Email = normalized.Email
		metadata["email"] = normalized.Email
	}
	if normalized.Language != nil {
		params.Language = normalized.Language
		metadata["language"] = *normalized.Language
	}
	if normalized.FirstName != nil {
		params.FirstName = normalized.FirstName
		metadata["firstName"] = normalized.FirstName
	}
	if normalized.LastName != nil {
		params.LastName = normalized.LastName
		metadata["lastName"] = normalized.LastName
	}
	if normalized.DisplayName != nil {
		params.DisplayName = normalized.DisplayName
		metadata["displayName"] = normalized.DisplayName
	}
	if normalized.PhoneNumber != nil {
		params.PhoneNumber = normalized.PhoneNumber
		metadata["phoneNumber"] = normalized.PhoneNumber
	}
	if normalized.WhatsappNumber != nil {
		params.WhatsappNumber = normalized.WhatsappNumber
		metadata["whatsappNumber"] = normalized.WhatsappNumber
	}
	if normalized.TelegramHandle != nil {
		params.TelegramHandle = normalized.TelegramHandle
		metadata["telegramHandle"] = normalized.TelegramHandle
	}
	if normalized.IsActive != nil {
		params.IsActive = normalized.IsActive
		metadata["isActive"] = *normalized.IsActive
	}

	if normalized.Password != nil {
		hash, hashErr := auth.HashPassword(*normalized.Password, s.passwordHashCost)
		if hashErr != nil {
			return UserRecord{}, hashErr
		}
		params.PasswordHash = &hash
	}

	updated, err := s.repo.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"user not found"}})
		}
		if mapped := mapConstraintError(err); mapped != nil {
			return UserRecord{}, mapped
		}
		return UserRecord{}, err
	}

	if normalized.Roles != nil {
		if roleErr := s.rbacService.SetUserRoles(ctx, in.UserID, *normalized.Roles); roleErr != nil {
			if errors.Is(roleErr, sql.ErrNoRows) {
				return UserRecord{}, apperror.ValidationWithDetails("validation failed", map[string]any{"roles": []string{"one or more roles are invalid"}})
			}
			return UserRecord{}, roleErr
		}
		metadata["roles"] = *normalized.Roles
	}

	roles, rolesErr := s.rbacService.RoleNamesForUser(ctx, updated.ID)
	if rolesErr != nil {
		return UserRecord{}, rolesErr
	}
	updated.Roles = roles

	if len(metadata) > 0 {
		s.logAudit(ctx, audit.Event{
			Action:      "users.update",
			ActorUserID: in.ActorID,
			EntityType:  "user",
			EntityID:    strPtr(fmt.Sprintf("%d", in.UserID)),
			Metadata:    metadata,
		})
	}

	s.rbacService.InvalidateUser(in.UserID)
	return updated, nil
}

func (s *Service) ResetPassword(ctx context.Context, actorID *int64, userID int64, newPassword string) error {
	hash, err := auth.HashPassword(newPassword, s.passwordHashCost)
	if err != nil {
		return err
	}
	if err := s.repo.SetPassword(ctx, userID, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ValidationWithDetails("validation failed", map[string]any{"id": []string{"user not found"}})
		}
		return err
	}
	s.logAudit(ctx, audit.Event{Action: "users.reset_password", ActorUserID: actorID, EntityType: "user", EntityID: strPtr(fmt.Sprintf("%d", userID))})
	return nil
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

func mapConstraintError(err error) *apperror.AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code != "23505" {
		return nil
	}
	if strings.Contains(pgErr.ConstraintName, "email") {
		return apperror.ValidationWithDetails("validation failed", map[string]any{"email": []string{"must be unique"}})
	}
	if strings.Contains(pgErr.ConstraintName, "username") {
		return apperror.ValidationWithDetails("validation failed", map[string]any{"username": []string{"must be unique"}})
	}
	return apperror.ValidationWithDetails("validation failed", map[string]any{"record": []string{"already exists"}})
}
