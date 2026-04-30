package cache

import (
	"context"
	"sync"
)

type LocalCache struct {
	mx      sync.RWMutex
	storage map[string][]byte
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		storage: make(map[string][]byte),
	}
}

func (c *LocalCache) Set(ctx context.Context, key string, val []byte) error {
	c.mx.Lock()
	c.storage[key] = val
	c.mx.Unlock()
	return nil
}

func (c *LocalCache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	res, ok := c.storage[key]
	return res, ok
}

func (c *LocalCache) Delete(ctx context.Context, key string) {
	c.mx.Lock()
	delete(c.storage, key)
	c.mx.Unlock()
}
