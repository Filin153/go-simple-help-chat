package usecase

import (
	"context"
	"shc/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type ScheduleRepo interface {
	Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error
	Update(ctx context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error
	Delete(ctx context.Context, id int, tx pgx.Tx) error
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

func (s *ScheduleUseCase) Edit(ctx context.Context, user domain.UserSystemInfo, editSchedule domain.EditSchedule) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}

	tx, err := s.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range editSchedule.Update {
		if err := s.scheduleRepo.Update(ctx, &item, tx); err != nil {
			return err
		}
	}

	for _, item := range editSchedule.Create {
		if err := s.scheduleRepo.Create(ctx, &item, tx); err != nil {
			return err
		}
	}

	for _, id := range editSchedule.Delete {
		if err := s.scheduleRepo.Delete(ctx, id, tx); err != nil {
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

	res := make([]domain.CreateSchedule, 0, 366)
	// now := time.Now().UTC()
	scheduleDate := time.Date(0000, 0, 0, 0, 0, 0, 0, time.UTC)
	for range 366 {
		day := scheduleDate.Weekday()

		item := domain.CreateSchedule{
			DepartmentID: departmentID,
			Name:         day.String(),
			WorkFrom:     scheduleDate.Add(time.Hour * 9),
			WorkTo:       scheduleDate.Add(time.Hour * 18),
			BreakFrom:    scheduleDate.Add(time.Hour * 13),
			BreakTo:      scheduleDate.Add(time.Hour * 14),
			IsWeekEnd:    false,
		}

		if day == time.Saturday || day == time.Sunday {
			item.IsWeekEnd = true
		}

		res = append(res, item)
		scheduleDate = scheduleDate.AddDate(0, 0, 1)
	}

	return res, nil
}
