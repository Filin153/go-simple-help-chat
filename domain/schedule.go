package domain

// ScheduleEventType describes what happens during a part of a schedule day.
type ScheduleEventType string

const (
	ScheduleEventWorkTime ScheduleEventType = "work_time"
	ScheduleEventBreak    ScheduleEventType = "break"
	ScheduleEventDayOff   ScheduleEventType = "day_off"
)

// CreateScheduleEvent uses a wall-clock time (HH:MM or HH:MM:SS). A day_off
// event has no time boundaries.
type CreateScheduleEvent struct {
	Type ScheduleEventType `json:"type" validate:"required,oneof=work_time break day_off"`
	From *string           `json:"from,omitempty"`
	To   *string           `json:"to,omitempty"`
}

// CreateScheduleEventRequest addresses the schedule day inside a batch edit.
type CreateScheduleEventRequest struct {
	ScheduleID int               `json:"schedule_id" validate:"required"`
	Type       ScheduleEventType `json:"type" validate:"required,oneof=work_time break day_off"`
	From       *string           `json:"from,omitempty"`
	To         *string           `json:"to,omitempty"`
}

type UpdateScheduleEvent struct {
	ID         int               `json:"id" validate:"required"`
	ScheduleID int               `json:"schedule_id" validate:"required"`
	Type       ScheduleEventType `json:"type" validate:"required,oneof=work_time break day_off"`
	From       *string           `json:"from,omitempty"`
	To         *string           `json:"to,omitempty"`
}

type DeleteScheduleEvent struct {
	ID         int `json:"id" validate:"required"`
	ScheduleID int `json:"schedule_id" validate:"required"`
}

type ScheduleEvent struct {
	ID         int               `json:"id"`
	ScheduleID int               `json:"schedule_id"`
	Type       ScheduleEventType `json:"type"`
	From       *string           `json:"from,omitempty"`
	To         *string           `json:"to,omitempty"`
}

// A schedule day is recurring: month/day identify it without tying it to a year.
type CreateSchedule struct {
	DepartmentID int                   `json:"department_id" validate:"required"`
	Month        int                   `json:"month" validate:"required,min=1,max=12"`
	Day          int                   `json:"day" validate:"required,min=1,max=31"`
	Name         string                `json:"name" validate:"required"`
	Events       []CreateScheduleEvent `json:"events" validate:"required,min=1,dive"`
}

// UpdateSchedule changes only events. The schedule day itself is immutable.
type UpdateSchedule struct {
	Create []CreateScheduleEventRequest `json:"create,omitempty" validate:"dive"`
	Update []UpdateScheduleEvent        `json:"update,omitempty" validate:"dive"`
	Delete []DeleteScheduleEvent        `json:"delete,omitempty" validate:"dive"`
}

type Schedule struct {
	ID           int             `json:"id"`
	DepartmentID int             `json:"department_id"`
	Month        int             `json:"month"`
	Day          int             `json:"day"`
	Name         string          `json:"name"`
	Events       []ScheduleEvent `json:"events"`
}
