package repository

import (
	"context"
	"shc/domain"
	"shc/internal/service"
	"time"

	"github.com/jackc/pgx/v5"
)

type ScheduleRepo struct {
	repo *Repository
}

func NewScheduleRepo(repo *Repository) *ScheduleRepo {
	return &ScheduleRepo{
		repo: repo,
	}
}

func (s *ScheduleRepo) Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
	const query = `INSERT INTO
	"schedules"("department_id", "name", "work_from", "work_to", "break_from", "break_to", "is_week_end")
	VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := s.repo.GetDB(tx).Exec(ctx, query, createSchedule.DepartmentID, createSchedule.Name, createSchedule.WorkFrom, createSchedule.WorkTo, createSchedule.BreakFrom, createSchedule.BreakTo, createSchedule.IsWeekEnd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ScheduleRepo) Update(ctx context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error {
	updateCol := service.StructToMap(updateSchedule, []string{"id"})
	query, args, err := getUpdateQuery("schedules", updateCol, map[string]any{
		"id": updateSchedule.ID,
	})
	if err != nil {
		return err
	}

	tag, err := s.repo.GetDB(tx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrZeroRowAffected
	}

	return nil
}

func (s *ScheduleRepo) Delete(ctx context.Context, id int, tx pgx.Tx) error {
	const query = `DELETE FROM "schedules" WHERE id=$1;`
	tag, err := s.repo.GetDB(tx).Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrZeroRowAffected
	}

	return nil
}

func (s *ScheduleRepo) ExistByDepartmentID(ctx context.Context, departmentID int) (bool, error) {
	var count int
	const query = `SELECT count("id") FROM "schedules" WHERE "department_id"=$1;`
	err := s.repo.GetDB(nil).QueryRow(ctx, query, departmentID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ScheduleRepo) GetFromTo(ctx context.Context, departmentID int, from, to time.Time, tx pgx.Tx) ([]domain.Schedule, error) {
	const query = `SELECT
	"id",
	"department_id",
	"name",
	"work_from",
	"work_to",
	"break_from",
	"break_to",
	"is_week_end"
FROM "schedules"
WHERE "department_id"=$1 AND "work_from" >= $2 AND "work_to" <= $3
ORDER BY "work_from" ASC, "id" ASC;`
	rows, err := s.repo.GetDB(tx).Query(ctx, query, departmentID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Schedule])
	if err != nil {
		return nil, err
	}

	return res, err
}
