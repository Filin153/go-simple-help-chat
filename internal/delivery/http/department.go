package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"shc/domain"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type DepartmentInterface interface {
	Create(ctx context.Context, user domain.UserSystemInfo, createDepartment domain.CreateDepartment) error
	GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.DepartmentWithOnScheduleeDay, error)
	GetByID(ctx context.Context, user domain.UserSystemInfo, id int) (*domain.Department, error)
	GetSheduleById(ctx context.Context, user domain.UserSystemInfo, id int, month int) ([]domain.Schedule, error)
	Update(ctx context.Context, user domain.UserSystemInfo, id int, updateDepartment domain.UpdateDepartment) error
	Delete(ctx context.Context, user domain.UserSystemInfo, id int) error
}

func (a *API) departmentCreate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	var createDepartment domain.CreateDepartment

	if err := json.NewDecoder(r.Body).Decode(&createDepartment); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()

	if err := a.validator.Struct(createDepartment); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	if err := a.department.Create(r.Context(), user, createDepartment); err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[any](w, 200, "OK", false, nil)
}

func (a *API) departmentGetAll(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		SendBaseResponse[any](w, 422, "page is not int", true, nil)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		SendBaseResponse[any](w, 422, "limit is not int", true, nil)
		return
	}

	res, err := a.department.GetAll(r.Context(), user, page, limit)
	if err == pgx.ErrNoRows {
		SendBaseResponse[any](w, 404, "", false, nil)
		return
	} else if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[[]domain.DepartmentWithOnScheduleeDay](w, 200, "OK", false, &res)
}

func (a *API) departmentGetByID(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	id := chi.URLParam(r, "id")
	if id == "" {
		SendBaseResponse[any](w, 422, "param id is nul", true, nil)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	res, err := a.department.GetByID(r.Context(), user, intID)
	if err == pgx.ErrNoRows {
		SendBaseResponse[any](w, 404, "", false, nil)
		return
	} else if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[domain.Department](w, 200, "OK", false, res)
}

func (a *API) departmentGetSheduleById(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	id := chi.URLParam(r, "id")
	if id == "" {
		SendBaseResponse[any](w, 422, "param id is nul", true, nil)
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	} else if month > 12 || month < 1 {
		SendBaseResponse[any](w, 422, "query param month is not correct. correct [1->12]", true, nil)
		return
	}

	res, err := a.department.GetSheduleById(r.Context(), user, intID, month)
	if err == pgx.ErrNoRows {
		SendBaseResponse[any](w, 404, "", false, nil)
		return
	} else if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[[]domain.Schedule](w, 200, "OK", false, &res)
}

func (a *API) departmentUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	id := chi.URLParam(r, "id")
	if id == "" {
		SendBaseResponse[any](w, 422, "param id is nul", true, nil)
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	var updateDepartment domain.UpdateDepartment
	if err := json.NewDecoder(r.Body).Decode(&updateDepartment); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()

	err = a.department.Update(r.Context(), user, intID, updateDepartment)
	if err == pgx.ErrNoRows || errors.Is(err, domain.ErrUnknownObject) {
		SendBaseResponse[any](w, 404, "", false, nil)
		return
	} else if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[any](w, 200, "OK", false, nil)
}

func (a *API) departmentDelete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	id := chi.URLParam(r, "id")
	if id == "" {
		SendBaseResponse[any](w, 422, "param id is nul", true, nil)
		return
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	err = a.department.Delete(r.Context(), user, intID)
	if err == pgx.ErrNoRows || errors.Is(err, domain.ErrUnknownObject) {
		SendBaseResponse[any](w, 404, "", false, nil)
		return
	} else if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[any](w, 200, "OK", false, nil)
}
