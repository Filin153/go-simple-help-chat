package cache

import (
	"context"
	"testing"
)

func Test_NewLocalCache(t *testing.T) {
	c := NewLocalCache()

	if c == nil {
		t.Fatal("NewLocalCache returned nil")
	}
	if c.storage == nil {
		t.Fatal("expected initialized storage map")
	}
}

func Test_LocalCache_SetGetDelete(t *testing.T) {
	c := NewLocalCache()
	ctx := context.Background()
	key := "user-1"
	val := []byte("hello")

	if err := c.Set(ctx, key, val); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if string(got) != string(val) {
		t.Fatalf("unexpected value: got=%q want=%q", string(got), string(val))
	}

	c.Delete(ctx, key)

	_, ok = c.Get(ctx, key)
	if ok {
		t.Fatal("expected key to be deleted")
	}
}
