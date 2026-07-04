package usecase

import (
	"context"
	"shc/domain"
)

type TicketRepo interface {
	Create(ctx context.Context, createTicket domain.CreateTicket) (int, error)
	GetAll(ctx context.Context, page, limit int) ([]domain.Ticket, error)
	GetAllByUserUUID(ctx context.Context, page, limit int, uuid string) ([]domain.Ticket, error)
}

type TicketUseCase struct {
	ticketRepo TicketRepo
}

func NewTicketUseCase(ticketRepo TicketRepo) *TicketUseCase {
	return &TicketUseCase{
		ticketRepo: ticketRepo,
	}
}

func (t *TicketUseCase) Create(ctx context.Context, user domain.UserSystemInfo) (int, error) {
	if user.UserRole != domain.UserRoleClient {
		return -1, domain.ErrAccess
	}

	return t.ticketRepo.Create(ctx, domain.CreateTicket{
		ClientUserUUID: user.UUID,
	})
}

func (t *TicketUseCase) GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.Ticket, error) {
	if user.UserRole != domain.UserRoleAdmin {
		return nil, domain.ErrAccess
	} else if limit > 100 {
		return nil, domain.ErrLimitIsBiggerThen100
	}

	return t.ticketRepo.GetAll(ctx, page, limit)
}

func (t *TicketUseCase) GetAllUser(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.Ticket, error) {
	if limit > 100 {
		return nil, domain.ErrLimitIsBiggerThen100
	}

	return t.ticketRepo.GetAllByUserUUID(ctx, page, limit, user.UUID)
}
