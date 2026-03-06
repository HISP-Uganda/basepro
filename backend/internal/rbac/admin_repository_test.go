package rbac

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestListPermissionsSupportsQueryAndModuleScope(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	repo := NewSQLRepository(sqlx.NewDb(sqlDB, "sqlmock"))
	module := "admin"
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM permissions p WHERE (p.name ILIKE $1 OR COALESCE(p.module_scope, '') ILIKE $1) AND p.module_scope = $2`)).
		WithArgs("%users%", module).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.id, p.name, p.module_scope, p.created_at, NULL::timestamptz AS assigned_at
		FROM permissions p
	 WHERE (p.name ILIKE $1 OR COALESCE(p.module_scope, '') ILIKE $1) AND p.module_scope = $2
		ORDER BY p.name ASC
		LIMIT $3 OFFSET $4`)).
		WithArgs("%users%", module, 25, 0).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "module_scope", "created_at", "assigned_at"}).
				AddRow(int64(1), "users.read", module, now, nil),
		)

	result, err := repo.ListPermissions(context.Background(), PermissionListQuery{
		Page:        1,
		PageSize:    25,
		SortField:   "name",
		SortOrder:   "asc",
		Query:       "users",
		ModuleScope: &module,
	})
	if err != nil {
		t.Fatalf("list permissions: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("expected one item, got total=%d len=%d", result.Total, len(result.Items))
	}
	if result.Items[0].Name != "users.read" {
		t.Fatalf("expected users.read, got %s", result.Items[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
