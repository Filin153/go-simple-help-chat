package service

import (
	"reflect"
	"testing"
	"time"
)

type structToMapFixture struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	WorkFrom  time.Time `json:"work_from"`
	IsWeekEnd bool      `json:"is_week_end"`
}

type structToMapTagPriorityFixture struct {
	Value string `db:"db_value" json:"json_value"`
}

func Test_StructToMap(t *testing.T) {
	workFrom := time.Date(2026, time.June, 18, 9, 0, 0, 0, time.UTC)

	got := StructToMap(&structToMapFixture{
		ID:       17,
		Name:     "Monday",
		WorkFrom: workFrom,
	}, []string{"id"})

	want := map[string]any{
		"name":      "Monday",
		"work_from": workFrom,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected map: got=%v want=%v", got, want)
	}
}

func Test_StructToMap_DBTagPriority(t *testing.T) {
	got := StructToMap(structToMapTagPriorityFixture{Value: "value"}, nil)

	want := map[string]any{
		"db_value": "value",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected map: got=%v want=%v", got, want)
	}
}
