package http

import (
	"context"
	"encoding/json"
	"net/http"
	"shc/domain"
)

type ScheduleInterface interface {
	Edit(ctx context.Context, user domain.UserSystemInfo, editSchedule domain.UpdateSchedule) error
}

func (a *API) scheduleEdit(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(domain.UserSystemInfo)

	var editSchedule domain.UpdateSchedule
	if err := json.NewDecoder(r.Body).Decode(&editSchedule); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}
	defer r.Body.Close()

	if err := a.validator.Struct(editSchedule); err != nil {
		SendBaseResponse[any](w, 422, err.Error(), true, nil)
		return
	}

	if err := a.schedule.Edit(r.Context(), user, editSchedule); err != nil {
		SendBaseResponse[any](w, 500, err.Error(), true, nil)
		return
	}

	SendBaseResponse[any](w, 200, "OK", false, nil)
}
