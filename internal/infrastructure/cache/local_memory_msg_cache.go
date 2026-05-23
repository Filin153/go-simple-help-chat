package cache

import (
	"context"
	"shc/domain"
	"sync"
)

type MsgCache struct {
	mx      sync.RWMutex
	storage map[string]map[int]*domain.Msg
}

func NewMsgCache() *MsgCache {
	return &MsgCache{
		storage: make(map[string]map[int]*domain.Msg),
	}
}

func (m *MsgCache) Set(ctx context.Context, userUUID string, val *domain.Msg) error {
	m.mx.Lock()
	if m.storage[userUUID] == nil {
		m.storage[userUUID] = make(map[int]*domain.Msg)
	}
	m.storage[userUUID][val.ID] = val
	m.mx.Unlock()
	return nil
}

func (m *MsgCache) Get(ctx context.Context, userUUID string) ([]*domain.Msg, bool, error) {
	m.mx.RLock()
	userMsgs := m.storage[userUUID]
	if len(userMsgs) == 0 {
		m.mx.RUnlock()
		return []*domain.Msg{}, false, nil
	}
	res := make([]*domain.Msg, 0, len(userMsgs))
	for _, val := range userMsgs {
		select {
		case <-ctx.Done():
			return []*domain.Msg{}, false, ctx.Err()
		default:
			res = append(res, val)
		}
	}
	m.mx.RUnlock()
	return res, true, nil
}

func (m *MsgCache) DeleteByMsgID(ctx context.Context, userUUID string, id int) error {
	m.mx.Lock()
	delete(m.storage[userUUID], id)
	m.mx.Unlock()
	return nil
}
