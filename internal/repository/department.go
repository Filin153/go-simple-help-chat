package repository

import (
	"context"
	"shc/domain"
	"shc/internal/service"

	"github.com/jackc/pgx/v5"
)

type DepartmentRepo struct {
	repo *Repository
}

func NewDepartmentRepo(repo *Repository) *DepartmentRepo {
	return &DepartmentRepo{
		repo: repo,
	}
}

func (d *DepartmentRepo) Create(ctx context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error) {
	const query = `INSERT INTO "departments"("name") VALUES ($1) RETURNING "id";`
	var res int
	err := d.repo.GetDB(tx).QueryRow(ctx, query, createDepartment.Name).Scan(&res)
	if err != nil {
		return -1, err
	}
	return res, nil
}

func (d *DepartmentRepo) GetAll(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error) {
	const query = `SELECT
	d."id",
	d."name",
	d."default_dep",
	s."work_from",
	s."work_to",
	s."is_week_end"
FROM "departments" AS d
JOIN LATERAL (
	SELECT
		"work_from",
		"work_to",
		"is_week_end"
	FROM "schedules"
	WHERE "department_id" = d."id"
	ORDER BY
		CASE
			WHEN EXTRACT(MONTH FROM "work_from") = EXTRACT(MONTH FROM CURRENT_DATE)
				AND EXTRACT(DAY FROM "work_from") = EXTRACT(DAY FROM CURRENT_DATE)
			THEN 0
			ELSE 1
		END,
		"work_from" ASC,
		"id" ASC
	LIMIT 1
) AS s ON TRUE
ORDER BY d."id" ASC
LIMIT $1 OFFSET $2;`
	rows, err := d.repo.GetDB(tx).Query(ctx, query, limit, getOffset(page, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DepartmentWithOnScheduleeDay])
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DepartmentRepo) GetByID(ctx context.Context, id int, tx pgx.Tx) (*domain.Department, error) {
	const query = `SELECT * FROM "departments" WHERE "id"=$1;`
	rows, err := d.repo.GetDB(tx).Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Department])
	if err != nil {
		return nil, err
	}

	return &res, err
}

func (d *DepartmentRepo) Update(ctx context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error {
	updateCol := service.StructToMap(updateDepartment, []string{})
	query, args, err := getUpdateQuery("departments", updateCol, map[string]any{
		"id": id,
	})
	if err != nil {
		return err
	}

	tag, err := d.repo.GetDB(tx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUnknownObject
	}
	return nil
}
