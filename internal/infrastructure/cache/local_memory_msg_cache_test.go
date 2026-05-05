package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"shc/domain"
)

func Test_NewMsgCache(t *testing.T) {
	c := NewMsgCache()

	if c == nil {
		t.Fatal("NewMsgCache returned nil")
	}
	if c.storage == nil {
		t.Fatal("expected initialized storage")
	}
}

func Test_MsgCache_SetGetDelete(t *testing.T) {
	c := NewMsgCache()
	ctx := context.Background()
	msg := &domain.Msg{ID: 10, Text: "hello"}

	if err := c.Set(ctx, "user-1", msg); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	got, ok, err := c.Get(ctx, "user-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(got) != 1 || got[0] != msg {
		t.Fatalf("unexpected messages: got=%v want=%v", got, []*domain.Msg{msg})
	}

	if err := c.DeleteByMsgID(ctx, "user-1", msg.ID); err != nil {
		t.Fatalf("DeleteByMsgID returned error: %v", err)
	}

	got, ok, err = c.Get(ctx, "user-1")
	if err != nil {
		t.Fatalf("Get after delete returned error: %v", err)
	}
	if ok {
		t.Fatal("expected ok to be false after delete")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result after delete, got=%v", got)
	}
}

func Test_MsgCache_Get_EmptyUser(t *testing.T) {
	c := NewMsgCache()

	got, ok, err := c.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if ok {
		t.Fatal("expected ok to be false for missing user")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got=%v", got)
	}
}

func Test_MsgCache_Get_ContextDone(t *testing.T) {
	c := NewMsgCache()
	if err := c.Set(context.Background(), "user-1", &domain.Msg{ID: 1}); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, ok, err := c.Get(ctx, "user-1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got=%v", err)
	}
	if ok {
		t.Fatal("expected ok to be false on canceled context")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got=%v", got)
	}
}

func Test_MsgCache_Set_MultipleMessages(t *testing.T) {
	c := NewMsgCache()
	msg1 := &domain.Msg{ID: 1, CreateAt: time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)}
	msg2 := &domain.Msg{ID: 2, CreateAt: time.Date(2026, 5, 5, 11, 0, 0, 0, time.UTC)}

	if err := c.Set(context.Background(), "user-1", msg1); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := c.Set(context.Background(), "user-1", msg2); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	got, ok, err := c.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got=%d", len(got))
	}
}
