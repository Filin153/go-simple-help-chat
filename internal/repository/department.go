package repository

import (
	"context"
	"fmt"
	"shc/domain"

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
	const query = `SELECT * FROM departments LIMIT $1 OFFSET $2;`
	rows, err := d.repo.GetDB(tx).Query(ctx, query, limit, getOffset(page, limit))
	if err != nil {
		return []domain.DepartmentWithOnScheduleeDay{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DepartmentWithOnScheduleeDay])
	if err != nil {
		return []domain.DepartmentWithOnScheduleeDay{}, err
	}
	return res, nil
}

func (d *DepartmentRepo) GetByID(ctx context.Context, id int, tx pgx.Tx) (domain.Department, error) {
	const query = `SELECT * FROM "departments" WHERE "id"=$1;`
	rows, err := d.repo.GetDB(tx).Query(ctx, query, id)
	if err != nil {
		return domain.Department{}, err
	}
	defer rows.Close()

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Department])
	if err != nil {
		return domain.Department{}, err
	}

	return res, err
}

func (d *DepartmentRepo) Update(ctx context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error {
	return fmt.Errorf("NEED TO DO")
}
