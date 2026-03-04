package users

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"basepro/backend/internal/apperror"
	"basepro/backend/internal/audit"
	"basepro/backend/internal/rbac"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

func TestListUsersPaginationIncludesTotalAndOffset(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	repo := NewSQLRepository(sqlx.NewDb(sqlDB, "sqlmock"))
	now := time.Now().UTC()
	email := "alice@example.com"
	display := "Alice Doe"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE username ILIKE $1 OR COALESCE(email, '') ILIKE $1 OR COALESCE(display_name, '') ILIKE $1`)).
		WithArgs("%ali%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, email, language, first_name, last_name, display_name,
		       phone_number, whatsapp_number, telegram_handle, is_active, last_login_at,
		       created_at, updated_at
		FROM users
	 WHERE username ILIKE $3 OR COALESCE(email, '') ILIKE $3 OR COALESCE(display_name, '') ILIKE $3 ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
		WithArgs(2, 2, "%ali%").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "email", "language", "first_name", "last_name", "display_name", "phone_number", "whatsapp_number", "telegram_handle", "is_active", "last_login_at", "created_at", "updated_at"}).
				AddRow(int64(3), "alice", email, "English", "Alice", "Doe", display, "+15551234567", nil, "alice_user", true, now, now, now).
				AddRow(int64(4), "alina", nil, "English", nil, nil, nil, nil, nil, nil, true, nil, now, now),
		)

	result, err := repo.ListUsers(context.Background(), ListQuery{
		Page:      2,
		PageSize:  2,
		SortField: "created_at",
		SortOrder: "desc",
		Filter:    "ali",
	})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if result.Total != 5 {
		t.Fatalf("expected total=5, got %d", result.Total)
	}
	if result.Page != 2 || result.PageSize != 2 {
		t.Fatalf("expected page metadata 2/2, got %d/%d", result.Page, result.PageSize)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Email == nil || *result.Items[0].Email != email {
		t.Fatalf("expected metadata email %q", email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

type fakeRepo struct {
	users          map[int64]UserRecord
	passwordHashes map[int64]string
	nextID         int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:          map[int64]UserRecord{},
		passwordHashes: map[int64]string{},
		nextID:         100,
	}
}

func (f *fakeRepo) ListUsers(_ context.Context, _ ListQuery) (ListResult, error) {
	items := make([]UserRecord, 0, len(f.users))
	for _, item := range f.users {
		items = append(items, item)
	}
	return ListResult{Items: items, Total: len(items), Page: 1, PageSize: 25}, nil
}

func (f *fakeRepo) GetUserByID(_ context.Context, userID int64) (UserRecord, error) {
	item, ok := f.users[userID]
	if !ok {
		return UserRecord{}, sql.ErrNoRows
	}
	return item, nil
}

func (f *fakeRepo) CreateUser(_ context.Context, params CreateUserParams) (UserRecord, error) {
	for _, user := range f.users {
		if strings.EqualFold(user.Username, params.Username) {
			return UserRecord{}, &pgconn.PgError{Code: "23505", ConstraintName: "users_username_key"}
		}
		if params.Email != nil && user.Email != nil && strings.EqualFold(*user.Email, *params.Email) {
			return UserRecord{}, &pgconn.PgError{Code: "23505", ConstraintName: "users_email_unique_idx"}
		}
	}

	f.nextID++
	now := time.Now().UTC()
	record := UserRecord{
		ID:             f.nextID,
		Username:       params.Username,
		Email:          params.Email,
		Language:       params.Language,
		FirstName:      params.FirstName,
		LastName:       params.LastName,
		DisplayName:    params.DisplayName,
		PhoneNumber:    params.PhoneNumber,
		WhatsappNumber: params.WhatsappNumber,
		TelegramHandle: params.TelegramHandle,
		IsActive:       params.IsActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		Roles:          []string{},
	}
	f.users[record.ID] = record
	f.passwordHashes[record.ID] = params.PasswordHash
	return record, nil
}

func (f *fakeRepo) UpdateUser(_ context.Context, params UpdateUserParams) (UserRecord, error) {
	record, ok := f.users[params.UserID]
	if !ok {
		return UserRecord{}, sql.ErrNoRows
	}
	if params.Username != nil {
		for _, user := range f.users {
			if user.ID != params.UserID && strings.EqualFold(user.Username, *params.Username) {
				return UserRecord{}, &pgconn.PgError{Code: "23505", ConstraintName: "users_username_key"}
			}
		}
		record.Username = *params.Username
	}
	if params.Email != nil {
		for _, user := range f.users {
			if user.ID != params.UserID && user.Email != nil && params.Email != nil && strings.EqualFold(*user.Email, *params.Email) {
				return UserRecord{}, &pgconn.PgError{Code: "23505", ConstraintName: "users_email_unique_idx"}
			}
		}
		record.Email = params.Email
	}
	if params.Language != nil {
		record.Language = *params.Language
	}
	if params.FirstName != nil {
		record.FirstName = params.FirstName
	}
	if params.LastName != nil {
		record.LastName = params.LastName
	}
	if params.DisplayName != nil {
		record.DisplayName = params.DisplayName
	}
	if params.PhoneNumber != nil {
		record.PhoneNumber = params.PhoneNumber
	}
	if params.WhatsappNumber != nil {
		record.WhatsappNumber = params.WhatsappNumber
	}
	if params.TelegramHandle != nil {
		record.TelegramHandle = params.TelegramHandle
	}
	if params.IsActive != nil {
		record.IsActive = *params.IsActive
	}
	if params.PasswordHash != nil {
		f.passwordHashes[params.UserID] = *params.PasswordHash
	}
	record.UpdatedAt = time.Now().UTC()
	f.users[params.UserID] = record
	return record, nil
}

func (f *fakeRepo) SetPassword(_ context.Context, userID int64, passwordHash string) error {
	if _, ok := f.users[userID]; !ok {
		return sql.ErrNoRows
	}
	f.passwordHashes[userID] = passwordHash
	return nil
}

type fakeRBACRepo struct {
	rolesByUser map[int64][]rbac.Role
}

func (f *fakeRBACRepo) GetUserRoles(_ context.Context, userID int64) ([]rbac.Role, error) {
	return append([]rbac.Role{}, f.rolesByUser[userID]...), nil
}

func (f *fakeRBACRepo) GetUserPermissions(context.Context, int64) ([]rbac.Permission, error) {
	return []rbac.Permission{}, nil
}

func (f *fakeRBACRepo) EnsureRole(context.Context, string) (rbac.Role, error) {
	return rbac.Role{}, nil
}

func (f *fakeRBACRepo) EnsurePermission(context.Context, string, *string) (rbac.Permission, error) {
	return rbac.Permission{}, nil
}

func (f *fakeRBACRepo) EnsureRolePermission(context.Context, int64, int64) error {
	return nil
}

func (f *fakeRBACRepo) EnsureUserRole(context.Context, int64, int64) error {
	return nil
}

func (f *fakeRBACRepo) GetRoleByName(_ context.Context, name string) (rbac.Role, error) {
	switch name {
	case "Viewer":
		return rbac.Role{ID: 4, Name: "Viewer"}, nil
	default:
		return rbac.Role{}, sql.ErrNoRows
	}
}

func (f *fakeRBACRepo) ReplaceUserRoles(_ context.Context, userID int64, roleIDs []int64) error {
	roles := make([]rbac.Role, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID == 4 {
			roles = append(roles, rbac.Role{ID: 4, Name: "Viewer"})
		}
	}
	f.rolesByUser[userID] = roles
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

func TestCreateUserWithMetadataReturnsIt(t *testing.T) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	auditRepo := &fakeAuditRepo{}
	service := NewService(repo, rbacService, audit.NewService(auditRepo), 4)

	email := "jane@example.com"
	firstName := "Jane"
	lastName := "Doe"
	phone := "+15550000001"
	whatsapp := "+15550000002"
	telegram := "jane_d"

	actorID := int64(1)
	created, err := service.CreateUser(context.Background(), CreateInput{
		Username:       "jane",
		Password:       "TempPass123!",
		Email:          &email,
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phone,
		WhatsappNumber: &whatsapp,
		TelegramHandle: &telegram,
		IsActive:       true,
		Roles:          []string{"Viewer"},
		ActorID:        &actorID,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if created.Email == nil || *created.Email != email {
		t.Fatalf("expected email %q", email)
	}
	if created.DisplayName == nil || *created.DisplayName != "Jane Doe" {
		t.Fatalf("expected derived display name Jane Doe, got %#v", created.DisplayName)
	}
	if len(auditRepo.events) == 0 || auditRepo.events[0].Action != "users.create" {
		t.Fatalf("expected users.create audit event")
	}
}

func TestUpdateUserChangesMetadataAndPasswordHash(t *testing.T) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	service := NewService(repo, rbacService, audit.NewService(&fakeAuditRepo{}), 4)

	email := "alpha@example.com"
	created, err := service.CreateUser(context.Background(), CreateInput{
		Username: "alpha",
		Password: "Pass123!",
		Email:    &email,
		IsActive: true,
		Roles:    []string{"Viewer"},
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	oldHash := repo.passwordHashes[created.ID]

	newEmail := "alpha.updated@example.com"
	password := "NewPass123!"
	display := "Alpha Updated"
	updated, err := service.UpdateUser(context.Background(), UpdateInput{
		UserID:      created.ID,
		Email:       &newEmail,
		DisplayName: &display,
		Password:    &password,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	if updated.Email == nil || *updated.Email != newEmail {
		t.Fatalf("expected updated email")
	}
	if updated.DisplayName == nil || *updated.DisplayName != display {
		t.Fatalf("expected updated display name")
	}
	if repo.passwordHashes[created.ID] == oldHash {
		t.Fatal("expected password hash to change")
	}
}

func TestListUsersReturnsMetadataFields(t *testing.T) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	service := NewService(repo, rbacService, audit.NewService(&fakeAuditRepo{}), 4)

	email := "listed@example.com"
	_, err := service.CreateUser(context.Background(), CreateInput{
		Username: "listed",
		Password: "Pass123!",
		Email:    &email,
		IsActive: true,
		Roles:    []string{"Viewer"},
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	list, err := service.ListUsers(context.Background(), ListQuery{Page: 1, PageSize: 25})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(list.Items))
	}
	if list.Items[0].Email == nil || *list.Items[0].Email != email {
		t.Fatalf("expected metadata email in list")
	}
}

func TestEmailUniquenessEnforced(t *testing.T) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	service := NewService(repo, rbacService, audit.NewService(&fakeAuditRepo{}), 4)

	email := "dup@example.com"
	_, err := service.CreateUser(context.Background(), CreateInput{Username: "first", Password: "Pass123!", Email: &email, IsActive: true})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	_, err = service.CreateUser(context.Background(), CreateInput{Username: "second", Password: "Pass123!", Email: &email, IsActive: true})
	if err == nil {
		t.Fatal("expected unique email validation error")
	}

	var appErr *apperror.AppError
	if !errorsAs(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperror.CodeValidationFailed {
		t.Fatalf("expected validation code, got %s", appErr.Code)
	}
}

func TestValidationErrorsUseStandardizedShapeCode(t *testing.T) {
	repo := newFakeRepo()
	rbacService := rbac.NewService(&fakeRBACRepo{rolesByUser: map[int64][]rbac.Role{}})
	service := NewService(repo, rbacService, audit.NewService(&fakeAuditRepo{}), 4)

	invalidEmail := "not-an-email"
	_, err := service.CreateUser(context.Background(), CreateInput{
		Username: "bad",
		Password: "Pass123!",
		Email:    &invalidEmail,
		IsActive: true,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apperror.AppError
	if !errorsAs(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperror.CodeValidationFailed {
		t.Fatalf("expected %s, got %s", apperror.CodeValidationFailed, appErr.Code)
	}
	if _, ok := appErr.Details["email"]; !ok {
		t.Fatalf("expected field-level details for email")
	}
}

func errorsAs(err error, target **apperror.AppError) bool {
	if err == nil {
		return false
	}
	appErr, ok := err.(*apperror.AppError)
	if ok {
		*target = appErr
		return true
	}
	return false
}
