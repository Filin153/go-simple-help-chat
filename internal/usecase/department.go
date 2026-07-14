package usecase

import (
	"context"
	"shc/domain"
	"time"

	"github.com/jackc/pgx/v5"
)

type DepartmentRepo interface {
	Create(ctx context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error)
	GetAll(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error)
	GetByID(ctx context.Context, id int, tx pgx.Tx) (*domain.Department, error)
	Update(ctx context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error
	Delete(ctx context.Context, id int) error
}

type ScheduleUseCaseInterface interface {
	GenerateBaseSchedule(ctx context.Context, user domain.UserSystemInfo, departmentID int) ([]domain.CreateSchedule, error)
}

type ScheduleRepoForDepartment interface {
	Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error
	GetFromTo(ctx context.Context, departmentID int, from, to time.Time, tx pgx.Tx) ([]domain.Schedule, error)
}

type DepartmentUseCase struct {
	mainRepo        MainRepo
	departmentRepo  DepartmentRepo
	scheduleUseCase ScheduleUseCaseInterface
	scheduleRepo    ScheduleRepoForDepartment
}

func NewDepartmentUseCase(mainRepo MainRepo, departmentRepo DepartmentRepo, scheduleUseCase ScheduleUseCaseInterface, scheduleRepo ScheduleRepoForDepartment) *DepartmentUseCase {
	return &DepartmentUseCase{
		mainRepo:        mainRepo,
		departmentRepo:  departmentRepo,
		scheduleUseCase: scheduleUseCase,
		scheduleRepo:    scheduleRepo,
	}
}

func (d *DepartmentUseCase) Create(ctx context.Context, user domain.UserSystemInfo, createDepartment domain.CreateDepartment) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}

	tx, err := d.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	depID, err := d.departmentRepo.Create(ctx, createDepartment, tx)
	if err != nil {
		return err
	}

	baseSchedule, err := d.scheduleUseCase.GenerateBaseSchedule(ctx, user, depID)
	if err != nil {
		return err
	}

	for _, item := range baseSchedule {
		if err := d.scheduleRepo.Create(ctx, &item, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (d *DepartmentUseCase) GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.DepartmentWithOnScheduleeDay, error) {
	if user.UserRole != domain.UserRoleAdmin {
		return nil, domain.ErrAccess
	}

	if limit > 100 {
		return nil, domain.ErrLimitIsBiggerThen100
	}
	return d.departmentRepo.GetAll(ctx, page, limit, nil)
}

func (d *DepartmentUseCase) GetByID(ctx context.Context, user domain.UserSystemInfo, id int) (*domain.Department, error) {
	return d.departmentRepo.GetByID(ctx, id, nil)
}

func (d *DepartmentUseCase) GetSheduleById(ctx context.Context, user domain.UserSystemInfo, id int, month int) ([]domain.Schedule, error) {
	from := time.Date(1, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(1, time.Month(month)+1, 0, 23, 59, 59, 999999999, time.UTC)
	return d.scheduleRepo.GetFromTo(ctx, id, from, to, nil)
}

func (d *DepartmentUseCase) Update(ctx context.Context, user domain.UserSystemInfo, id int, updateDepartment domain.UpdateDepartment) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}
	return d.departmentRepo.Update(ctx, id, updateDepartment, nil)
}

func (d *DepartmentUseCase) Delete(ctx context.Context, user domain.UserSystemInfo, id int) error {
	if user.UserRole != domain.UserRoleAdmin {
		return domain.ErrAccess
	}
	return d.departmentRepo.Delete(ctx, id)
}
