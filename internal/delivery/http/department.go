package http

import (
	"context"
	"encoding/json"
	"net/http"
	"shc/domain"
	"strconv"
)

type DepartmentInterface interface {
	Create(ctx context.Context, user domain.UserSystemInfo, createDepartment domain.CreateDepartment) error
	GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int) ([]domain.DepartmentWithOnScheduleeDay, error)
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
	if err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[[]domain.DepartmentWithOnScheduleeDay](w, 200, "OK", false, &res)
}
