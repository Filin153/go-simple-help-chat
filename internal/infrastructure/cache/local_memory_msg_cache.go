package cache

import (
	"context"
	"shc/domain"
	"sync"
)

type MsgCache struct {
	userTiketMX sync.RWMutex
	msgStoreMX  sync.RWMutex
	userTikets  map[string]map[int]struct{}
	msgStore    map[int]map[int]domain.Msg
}

func NewMsgCache() *MsgCache {
	return &MsgCache{
		userTikets: make(map[string]map[int]struct{}),
		msgStore:   make(map[int]map[int]domain.Msg),
	}
}

func (m *MsgCache) SetMsg(ctx context.Context, msg *domain.Msg) error {
	m.userTiketMX.Lock()
	if _, ok := m.userTikets[msg.UserUUID]; !ok {
		m.userTikets[msg.UserUUID] = make(map[int]struct{})
	}
	m.userTikets[msg.UserUUID][msg.TicketID] = struct{}{}
	m.userTiketMX.Unlock()

	m.msgStoreMX.Lock()
	if _, ok := m.msgStore[msg.TicketID]; !ok {
		m.msgStore[msg.TicketID] = make(map[int]domain.Msg)
	}
	m.msgStore[msg.TicketID][msg.ID] = *msg
	m.msgStoreMX.Unlock()

	return nil
}

func (m *MsgCache) GetMsgByUserUUID(ctx context.Context, userUUID string) ([]domain.Msg, bool, error) {
	tikets, ok := m.getUserTikets(userUUID)
	res := make([]domain.Msg, 0, len(tikets))

	if len(tikets) == 0 || !ok {
		return res, false, nil
	}

	for _, tiketID := range tikets {
		msges, _, err := m.GetMsgByTiketID(ctx, tiketID)
		if err != nil {
			return res, false, err
		}
		res = append(res, msges...)
	}

	return res, true, nil
}
func (m *MsgCache) GetMsgByTiketID(ctx context.Context, ticketID int) ([]domain.Msg, bool, error) {
	res := make([]domain.Msg, 0, 1)

	m.msgStoreMX.RLock()
	tiketMsges, ok := m.msgStore[ticketID]
	m.msgStoreMX.RUnlock()

	for _, msg := range tiketMsges {
		res = append(res, msg)
	}

	if len(res) == 0 || !ok {
		return res, false, nil
	}

	return res, true, nil
}

func (m *MsgCache) DeleteByMsgID(ctx context.Context, userUUID string, msgIDs []int) error {
	tikets, _ := m.getUserTikets(userUUID)

	m.msgStoreMX.Lock()
	defer m.msgStoreMX.Unlock()

	for _, tiketID := range tikets {
		for _, msgID := range msgIDs {
			delete(m.msgStore[tiketID], msgID)
		}

		if len(m.msgStore[tiketID]) == 0 {
			m.userTiketMX.Lock()
			delete(m.userTikets[userUUID], tiketID)
			m.userTiketMX.Unlock()
		}
	}

	return nil
}

func (m *MsgCache) getUserTikets(userUUID string) ([]int, bool) {
	m.userTiketMX.RLock()
	userT, ok := m.userTikets[userUUID]
	m.userTiketMX.RUnlock()

	res := make([]int, 0, len(userT))
	for id, _ := range userT {
		res = append(res, id)
	}
	return res, ok
}
