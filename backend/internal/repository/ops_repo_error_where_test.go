package repository

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestBuildOpsErrorLogsWhere_QueryUsesQualifiedColumns(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		Query: "ACCESS_DENIED",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "e.request_id ILIKE $") {
		t.Fatalf("where should include qualified request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.client_request_id ILIKE $") {
		t.Fatalf("where should include qualified client_request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.error_message ILIKE $") {
		t.Fatalf("where should include qualified error_message condition: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_UserQueryUsesExistsSubquery(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		UserQuery: "admin@",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id AND u.email ILIKE $") {
		t.Fatalf("where should include EXISTS user email condition: %s", where)
	}
}

func TestOpsRepositoryDeleteErrorLogsUsesSharedWhere(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ops_error_logs e WHERE 1=1 AND (COALESCE(e.status_code, 0) >= 400 OR e.error_type = 'cyber_policy') AND e.created_at >= $1 AND e.created_at < $2 AND COALESCE(e.is_business_limited,false) = false AND COALESCE(e.requested_model, e.model, '') = $3")).
		WithArgs(start, end, "gpt-5.3-codex").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectClose()

	deleted, err := repo.DeleteErrorLogs(context.Background(), &service.OpsErrorLogFilter{
		StartTime: &start,
		EndTime:   &end,
		Model:     "gpt-5.3-codex",
	})
	if err != nil {
		t.Fatalf("DeleteErrorLogs() error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted=%d, want 2", deleted)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("db.Close() error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
