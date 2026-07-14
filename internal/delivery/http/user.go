package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"shc/domain"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UserInterface interface {
	GetAll(ctx context.Context, user domain.UserSystemInfo, page, limit int, filter domain.UserFilter) ([]domain.User, error)
	GetByUUID(ctx context.Context, user domain.UserSystemInfo, uuid string) (*domain.User, error)
	GetByLoginForAdmin(ctx context.Context, user domain.UserSystemInfo, login string) (*domain.User, error)
	CreateManager(ctx context.Context, user domain.UserSystemInfo, manager domain.CreateManager) error
	CreateAdmin(ctx context.Context, user domain.UserSystemInfo, admin domain.CreateUser) error
	UpdateByUUID(ctx context.Context, user domain.UserSystemInfo, uuid string, userForUpdate domain.UpdateUser) error
	DeleteByUUID(ctx context.Context, user domain.UserSystemInfo, uuid string) error
}

func sendUserMutationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAccess):
		SendBaseResponse[any](w, http.StatusForbidden, err.Error(), true, nil)
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, domain.ErrUnknownObject):
		SendBaseResponse[any](w, http.StatusNotFound, err.Error(), true, nil)
	case errors.Is(err, domain.ErrShortPassword), errors.Is(err, domain.ErrEmptyObject):
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
	default:
		SendBaseResponse[any](w, http.StatusInternalServerError, err.Error(), true, nil)
	}
}

func (a *API) userGetByUUID(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)
	res, err := a.user.GetByUUID(r.Context(), user, chi.URLParam(r, "uuid"))
	if err != nil {
		sendUserMutationError(w, err)
		return
	}
	SendBaseResponse[domain.User](w, http.StatusOK, "OK", false, res)
}

func (a *API) userCreateManager(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)
	var manager domain.CreateManager
	if err := json.NewDecoder(r.Body).Decode(&manager); err != nil {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()
	if err := a.validator.Struct(manager); err != nil {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	if err := a.user.CreateManager(r.Context(), user, manager); err != nil {
		sendUserMutationError(w, err)
		return
	}
	SendBaseResponse[any](w, http.StatusCreated, "OK", false, nil)
}

func (a *API) userCreateAdmin(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)
	var admin domain.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()
	if err := a.validator.Struct(admin); err != nil {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	if err := a.user.CreateAdmin(r.Context(), user, admin); err != nil {
		sendUserMutationError(w, err)
		return
	}
	SendBaseResponse[any](w, http.StatusCreated, "OK", false, nil)
}

func (a *API) userUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)
	var update domain.UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()
	if err := a.user.UpdateByUUID(r.Context(), user, chi.URLParam(r, "uuid"), update); err != nil {
		sendUserMutationError(w, err)
		return
	}
	SendBaseResponse[any](w, http.StatusOK, "OK", false, nil)
}

func (a *API) userDelete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)
	if err := a.user.DeleteByUUID(r.Context(), user, chi.URLParam(r, "uuid")); err != nil {
		sendUserMutationError(w, err)
		return
	}
	SendBaseResponse[any](w, http.StatusOK, "OK", false, nil)
}

func (a *API) userGetAll(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	page, limit := 1, 20
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 {
			SendBaseResponse[any](w, http.StatusUnprocessableEntity, "page должен быть положительным целым числом", true, nil)
			return
		}
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 {
			SendBaseResponse[any](w, http.StatusUnprocessableEntity, "limit должен быть положительным целым числом", true, nil)
			return
		}
	}

	filter := domain.UserFilter{
		Role:   domain.UserRole(strings.TrimSpace(r.URL.Query().Get("role"))),
		Search: strings.TrimSpace(r.URL.Query().Get("search")),
	}
	res, err := a.user.GetAll(r.Context(), user, page, limit, filter)
	if errors.Is(err, domain.ErrAccess) {
		SendBaseResponse[any](w, http.StatusForbidden, err.Error(), true, nil)
		return
	}
	if errors.Is(err, domain.ErrInvalidUserRole) || errors.Is(err, domain.ErrLimitIsBiggerThen100) {
		SendBaseResponse[any](w, http.StatusUnprocessableEntity, err.Error(), true, nil)
		return
	}
	if err != nil {
		SendBaseResponse[any](w, http.StatusInternalServerError, err.Error(), true, nil)
		return
	}

	SendBaseResponse[[]domain.User](w, http.StatusOK, "OK", false, &res)
}
