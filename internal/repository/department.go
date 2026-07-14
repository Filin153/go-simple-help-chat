package repository

import (
	"context"
	"encoding/json"
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
	const query = `SELECT d.id,d.name,d.default_dep,
		CASE WHEN s.id IS NULL THEN NULL ELSE jsonb_build_object(
			'id',s.id,'department_id',s.department_id,'month',s.month,'day',s.day,'name',s.name,
			'events',COALESCE((SELECT jsonb_agg(jsonb_build_object(
				'id',e.id,'schedule_id',e.schedule_id,'type',e.type,
				'from',e.time_from::text,'to',e.time_to::text) ORDER BY e.time_from NULLS FIRST,e.id)
				FROM schedule_events e WHERE e.schedule_id=s.id),'[]'::jsonb)) END
	FROM departments d
	LEFT JOIN schedules s ON s.department_id=d.id
		AND s.month=EXTRACT(MONTH FROM CURRENT_DATE)::int AND s.day=EXTRACT(DAY FROM CURRENT_DATE)::int
	ORDER BY d.id LIMIT $1 OFFSET $2`
	rows, err := d.repo.GetDB(tx).Query(ctx, query, limit, getOffset(page, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.DepartmentWithOnScheduleeDay, 0)
	for rows.Next() {
		var item domain.DepartmentWithOnScheduleeDay
		var raw []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Default, &raw); err != nil {
			return nil, err
		}
		if raw != nil {
			item.Schedule = &domain.Schedule{}
			if err := json.Unmarshal(raw, item.Schedule); err != nil {
				return nil, err
			}
		}
		res = append(res, item)
	}
	return res, rows.Err()
}

func (d *DepartmentRepo) GetByID(ctx context.Context, id int, tx pgx.Tx) (*domain.Department, error) {
	const query = `SELECT "id", "name", "default_dep" FROM "departments" WHERE "id"=$1;`
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

func (d *DepartmentRepo) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM "departments" WHERE "id"=$1`
	_, err := d.repo.GetDB(nil).Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
