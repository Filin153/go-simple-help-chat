package repository

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"shc/domain"
)

func Test_GetUpdateQuery_EmptyObject(t *testing.T) {
	query, args, err := getUpdateQuery("schedules", map[string]any{}, map[string]any{"id": 1})
	if !errors.Is(err, domain.ErrEmptyObject) {
		t.Fatalf("expected ErrEmptyObject, got=%v", err)
	}
	if query != "" {
		t.Fatalf("expected empty query, got=%q", query)
	}
	if args != nil {
		t.Fatalf("expected nil args, got=%v", args)
	}
}

func Test_GetUpdateQuery_OK(t *testing.T) {
	workFrom := time.Date(2026, time.June, 18, 9, 0, 0, 0, time.UTC)

	query, args, err := getUpdateQuery("schedules", map[string]any{
		"work_from": workFrom,
		"name":      "Monday",
	}, map[string]any{
		"id": 17,
	})
	if err != nil {
		t.Fatalf("getUpdateQuery returned error: %v", err)
	}

	wantQuery := `UPDATE "schedules" SET "name" = $1, "work_from" = $2 WHERE "id" = $3;`
	if query != wantQuery {
		t.Fatalf("unexpected query: got=%q want=%q", query, wantQuery)
	}

	wantArgs := []any{"Monday", workFrom, 17}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("unexpected args: got=%v want=%v", args, wantArgs)
	}
}
