package repository

import (
	"context"
	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type ScheduleRepo struct{ repo *Repository }

func NewScheduleRepo(repo *Repository) *ScheduleRepo { return &ScheduleRepo{repo: repo} }

func (s *ScheduleRepo) Create(ctx context.Context, item *domain.CreateSchedule, tx pgx.Tx) error {
	const dayQuery = `INSERT INTO schedules (department_id, month, day, name) VALUES ($1,$2,$3,$4) RETURNING id`
	var scheduleID int
	if err := s.repo.GetDB(tx).QueryRow(ctx, dayQuery, item.DepartmentID, item.Month, item.Day, item.Name).Scan(&scheduleID); err != nil {
		return err
	}
	return s.createEvents(ctx, scheduleID, item.Events, tx)
}

func (s *ScheduleRepo) createEvents(ctx context.Context, scheduleID int, events []domain.CreateScheduleEvent, tx pgx.Tx) error {
	for _, event := range events {
		if err := s.CreateEvent(ctx, scheduleID, &event, tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *ScheduleRepo) CreateEvent(ctx context.Context, scheduleID int, event *domain.CreateScheduleEvent, tx pgx.Tx) error {
	const query = `INSERT INTO schedule_events (schedule_id,type,time_from,time_to)
		VALUES ($1,$2,CASE WHEN $2::schedule_event_type='day_off' THEN NULL ELSE $3::time END,
		CASE WHEN $2::schedule_event_type='day_off' THEN NULL ELSE $4::time END)`
	_, err := s.repo.GetDB(tx).Exec(ctx, query, scheduleID, event.Type, event.From, event.To)
	return err
}

func (s *ScheduleRepo) UpdateEvent(ctx context.Context, scheduleID int, event *domain.UpdateScheduleEvent, tx pgx.Tx) error {
	const query = `UPDATE schedule_events SET type=$3,
		time_from=CASE WHEN $3::schedule_event_type='day_off' THEN NULL ELSE $4::time END,
		time_to=CASE WHEN $3::schedule_event_type='day_off' THEN NULL ELSE $5::time END
		WHERE id=$1 AND schedule_id=$2`
	tag, err := s.repo.GetDB(tx).Exec(ctx, query, event.ID, scheduleID, event.Type, event.From, event.To)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrZeroRowAffected
	}
	return nil
}

func (s *ScheduleRepo) DeleteEvent(ctx context.Context, scheduleID, eventID int, tx pgx.Tx) error {
	tag, err := s.repo.GetDB(tx).Exec(ctx, `DELETE FROM schedule_events WHERE id=$1 AND schedule_id=$2`, eventID, scheduleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrZeroRowAffected
	}
	return nil
}

func (s *ScheduleRepo) ExistByDepartmentID(ctx context.Context, departmentID int) (bool, error) {
	var exists bool
	err := s.repo.GetDB(nil).QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schedules WHERE department_id=$1)`, departmentID).Scan(&exists)
	return exists, err
}

func (s *ScheduleRepo) GetByMonth(ctx context.Context, departmentID, month int, tx pgx.Tx) ([]domain.Schedule, error) {
	const query = `SELECT s.id,s.department_id,s.month,s.day,s.name,e.id,e.type::text,e.time_from::text,e.time_to::text
		FROM schedules s LEFT JOIN schedule_events e ON e.schedule_id=s.id
		WHERE s.department_id=$1 AND s.month=$2 ORDER BY s.day,s.id,e.time_from NULLS FIRST,e.id`
	rows, err := s.repo.GetDB(tx).Query(ctx, query, departmentID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Schedule, 0)
	index := make(map[int]int)
	for rows.Next() {
		var day domain.Schedule
		var eventID *int
		var eventType, from, to *string
		if err := rows.Scan(&day.ID, &day.DepartmentID, &day.Month, &day.Day, &day.Name, &eventID, &eventType, &from, &to); err != nil {
			return nil, err
		}
		pos, ok := index[day.ID]
		if !ok {
			day.Events = []domain.ScheduleEvent{}
			result = append(result, day)
			pos = len(result) - 1
			index[day.ID] = pos
		}
		if eventID != nil {
			result[pos].Events = append(result[pos].Events, domain.ScheduleEvent{
				ID: *eventID, ScheduleID: day.ID, Type: domain.ScheduleEventType(*eventType), From: from, To: to,
			})
		}
	}
	return result, rows.Err()
}
