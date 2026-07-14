package usecase

import (
	"context"
	"shc/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type ScheduleRepo interface {
	Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error
	CreateEvent(ctx context.Context, scheduleID int, event *domain.CreateScheduleEvent, tx pgx.Tx) error
	UpdateEvent(ctx context.Context, scheduleID int, event *domain.UpdateScheduleEvent, tx pgx.Tx) error
	DeleteEvent(ctx context.Context, scheduleID, eventID int, tx pgx.Tx) error
	ExistByDepartmentID(ctx context.Context, departmentID int) (bool, error)
}

type ScheduleUseCase struct {
	mainRepo     MainRepo
	scheduleRepo ScheduleRepo
}

func NewScheduleUseCase(mainRepo MainRepo, scheduleRepo ScheduleRepo) *ScheduleUseCase {
	return &ScheduleUseCase{
		mainRepo:     mainRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *ScheduleUseCase) Edit(ctx context.Context, user domain.UserSystemInfo, editSchedule domain.UpdateSchedule) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}
	if len(editSchedule.Create) == 0 && len(editSchedule.Update) == 0 && len(editSchedule.Delete) == 0 {
		return domain.ErrEmptyObject
	}

	tx, err := s.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range editSchedule.Create {
		event := editSchedule.Create[i]
		createEvent := domain.CreateScheduleEvent{Type: event.Type, From: event.From, To: event.To}
		if err := s.scheduleRepo.CreateEvent(ctx, event.ScheduleID, &createEvent, tx); err != nil {
			return err
		}
	}
	for i := range editSchedule.Update {
		event := &editSchedule.Update[i]
		if err := s.scheduleRepo.UpdateEvent(ctx, event.ScheduleID, event, tx); err != nil {
			return err
		}
	}
	for _, event := range editSchedule.Delete {
		if err := s.scheduleRepo.DeleteEvent(ctx, event.ScheduleID, event.ID, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *ScheduleUseCase) GenerateBaseSchedule(ctx context.Context, user domain.UserSystemInfo, departmentID int) ([]domain.CreateSchedule, error) {
	if user.UserRole != domain.UserRoleAdmin {
		return nil, domain.ErrAccess
	}

	exist, err := s.scheduleRepo.ExistByDepartmentID(ctx, departmentID)
	if err != nil {
		return []domain.CreateSchedule{}, err
	} else if exist {
		return []domain.CreateSchedule{}, nil
	}

	res := make([]domain.CreateSchedule, 0, 365)
	scheduleDate := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	for scheduleDate.Year() == 1 {
		day := scheduleDate.Weekday()
		workFrom, workTo := "09:00:00", "18:00:00"
		breakFrom, breakTo := "13:00:00", "14:00:00"
		item := domain.CreateSchedule{
			DepartmentID: departmentID,
			Month:        int(scheduleDate.Month()),
			Day:          scheduleDate.Day(),
			Name:         day.String(),
			Events: []domain.CreateScheduleEvent{
				{Type: domain.ScheduleEventWorkTime, From: &workFrom, To: &workTo},
				{Type: domain.ScheduleEventBreak, From: &breakFrom, To: &breakTo},
			},
		}

		if day == time.Saturday || day == time.Sunday {
			item.Events = []domain.CreateScheduleEvent{{Type: domain.ScheduleEventDayOff}}
		}

		res = append(res, item)
		scheduleDate = scheduleDate.AddDate(0, 0, 1)
	}

	return res, nil
}
