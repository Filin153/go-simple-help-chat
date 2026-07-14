package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type scheduleRepoMock struct {
	createFn              func(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error
	updateFn              func(ctx context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error
	deleteFn              func(ctx context.Context, id int, tx pgx.Tx) error
	existByDepartmentIDFn func(ctx context.Context, departmentID int) (bool, error)
}

func (s *scheduleRepoMock) Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
	return s.createFn(ctx, createSchedule, tx)
}

func (s *scheduleRepoMock) Update(ctx context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error {
	return s.updateFn(ctx, updateSchedule, tx)
}

func (s *scheduleRepoMock) Delete(ctx context.Context, id int, tx pgx.Tx) error {
	return s.deleteFn(ctx, id, tx)
}

func (s *scheduleRepoMock) ExistByDepartmentID(ctx context.Context, departmentID int) (bool, error) {
	return s.existByDepartmentIDFn(ctx, departmentID)
}

type scheduleFixture struct {
	useCase      *ScheduleUseCase
	tx           *fakeTx
	mainRepo     *mainRepoMock
	scheduleRepo *scheduleRepoMock
}

func newScheduleFixture() *scheduleFixture {
	tx := &fakeTx{}
	mainRepo := &mainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return tx, nil
		},
	}
	scheduleRepo := &scheduleRepoMock{
		createFn: func(_ context.Context, _ *domain.CreateSchedule, _ pgx.Tx) error {
			return nil
		},
		updateFn: func(_ context.Context, _ *domain.UpdateSchedule, _ pgx.Tx) error {
			return nil
		},
		deleteFn: func(_ context.Context, _ int, _ pgx.Tx) error {
			return nil
		},
		existByDepartmentIDFn: func(_ context.Context, _ int) (bool, error) {
			return false, nil
		},
	}

	return &scheduleFixture{
		useCase:      NewScheduleUseCase(mainRepo, scheduleRepo),
		tx:           tx,
		mainRepo:     mainRepo,
		scheduleRepo: scheduleRepo,
	}
}

func Test_NewScheduleUseCase(t *testing.T) {
	f := newScheduleFixture()

	if f.useCase == nil {
		t.Fatal("NewScheduleUseCase returned nil")
	}
	if f.useCase.mainRepo != f.mainRepo {
		t.Fatal("unexpected mainRepo in ScheduleUseCase")
	}
	if f.useCase.scheduleRepo != f.scheduleRepo {
		t.Fatal("unexpected scheduleRepo in ScheduleUseCase")
	}
}

func Test_ScheduleUseCase_Edit_CreateSessionError(t *testing.T) {
	f := newScheduleFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
}

func Test_ScheduleUseCase_Edit_UpdateError(t *testing.T) {
	f := newScheduleFixture()
	wantErr := errors.New("update error")
	f.scheduleRepo.updateFn = func(_ context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Update")
		}
		if updateSchedule.ID != 10 {
			t.Fatalf("unexpected update schedule id: got=%d want=10", updateSchedule.ID)
		}
		return wantErr
	}

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{
		Update: []domain.UpdateSchedule{{ID: 10, Name: "monday"}},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected update error, got=%v", err)
	}
}

func Test_ScheduleUseCase_Edit_CreateError(t *testing.T) {
	f := newScheduleFixture()
	updateCalled := false
	wantErr := errors.New("create error")
	f.scheduleRepo.updateFn = func(_ context.Context, _ *domain.UpdateSchedule, tx pgx.Tx) error {
		updateCalled = true
		if tx != f.tx {
			t.Fatal("unexpected tx in Update")
		}
		return nil
	}
	f.scheduleRepo.createFn = func(_ context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Create")
		}
		if createSchedule.DepartmentID != 5 {
			t.Fatalf("unexpected department id: got=%d want=5", createSchedule.DepartmentID)
		}
		return wantErr
	}

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{
		Update: []domain.UpdateSchedule{{ID: 1}},
		Create: []domain.CreateSchedule{{DepartmentID: 5}},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create error, got=%v", err)
	}
	if !updateCalled {
		t.Fatal("expected Update to be called before Create")
	}
}

func Test_ScheduleUseCase_Edit_DeleteError(t *testing.T) {
	f := newScheduleFixture()
	createCalled := false
	wantErr := errors.New("delete error")
	f.scheduleRepo.createFn = func(_ context.Context, _ *domain.CreateSchedule, tx pgx.Tx) error {
		createCalled = true
		if tx != f.tx {
			t.Fatal("unexpected tx in Create")
		}
		return nil
	}
	f.scheduleRepo.deleteFn = func(_ context.Context, id int, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Delete")
		}
		if id != 7 {
			t.Fatalf("unexpected delete id: got=%d want=7", id)
		}
		return wantErr
	}

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{
		Create: []domain.CreateSchedule{{DepartmentID: 1}},
		Delete: []int{7},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete error, got=%v", err)
	}
	if !createCalled {
		t.Fatal("expected Create to be called before Delete")
	}
}

func Test_ScheduleUseCase_Edit_CommitError(t *testing.T) {
	f := newScheduleFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_ScheduleUseCase_Edit_OK(t *testing.T) {
	f := newScheduleFixture()
	callOrder := make([]string, 0, 3)
	f.scheduleRepo.updateFn = func(_ context.Context, updateSchedule *domain.UpdateSchedule, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Update")
		}
		callOrder = append(callOrder, updateSchedule.Name)
		return nil
	}
	f.scheduleRepo.createFn = func(_ context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Create")
		}
		callOrder = append(callOrder, createSchedule.Name)
		return nil
	}
	f.scheduleRepo.deleteFn = func(_ context.Context, id int, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("unexpected tx in Delete")
		}
		callOrder = append(callOrder, "delete")
		return nil
	}

	err := f.useCase.Edit(context.Background(), []domain.UpdateSchedule{
		Update: []domain.UpdateSchedule{{ID: 1, Name: "update"}},
		Create: []domain.CreateSchedule{{DepartmentID: 2, Name: "create"}},
		Delete: []int{3},
	})
	if err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}
	if len(callOrder) != 3 {
		t.Fatalf("unexpected call count: got=%d want=3", len(callOrder))
	}
	if callOrder[0] != "update" || callOrder[1] != "create" || callOrder[2] != "delete" {
		t.Fatalf("unexpected call order: got=%v", callOrder)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit once, got=%d", f.tx.commitCalls)
	}
}

func Test_ScheduleUseCase_GenerateBaseSchedule_ExistByDepartmentIDError(t *testing.T) {
	f := newScheduleFixture()
	wantErr := errors.New("exist error")
	f.scheduleRepo.existByDepartmentIDFn = func(_ context.Context, departmentID int) (bool, error) {
		if departmentID != 9 {
			t.Fatalf("unexpected department id: got=%d want=9", departmentID)
		}
		return false, wantErr
	}

	items, err := f.useCase.GenerateBaseSchedule(context.Background(), 9)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected exist error, got=%v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty result on error, got=%d items", len(items))
	}
}

func Test_ScheduleUseCase_GenerateBaseSchedule_AlreadyExists(t *testing.T) {
	f := newScheduleFixture()
	f.scheduleRepo.existByDepartmentIDFn = func(_ context.Context, departmentID int) (bool, error) {
		if departmentID != 11 {
			t.Fatalf("unexpected department id: got=%d want=11", departmentID)
		}
		return true, nil
	}

	items, err := f.useCase.GenerateBaseSchedule(context.Background(), 11)
	if err != nil {
		t.Fatalf("GenerateBaseSchedule returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty result when schedule exists, got=%d items", len(items))
	}
}

func Test_ScheduleUseCase_GenerateBaseSchedule_OK(t *testing.T) {
	f := newScheduleFixture()

	items, err := f.useCase.GenerateBaseSchedule(context.Background(), 15)
	if err != nil {
		t.Fatalf("GenerateBaseSchedule returned error: %v", err)
	}
	if len(items) != 365 {
		t.Fatalf("unexpected schedule len: got=%d want=365", len(items))
	}

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	first := items[0]
	if first.DepartmentID != 15 {
		t.Fatalf("unexpected department id: got=%d want=15", first.DepartmentID)
	}
	if first.Name != start.Weekday().String() {
		t.Fatalf("unexpected first day name: got=%q want=%q", first.Name, start.Weekday().String())
	}
	if !first.WorkFrom.Equal(start.Add(9 * time.Hour)) {
		t.Fatalf("unexpected WorkFrom: got=%v want=%v", first.WorkFrom, start.Add(9*time.Hour))
	}
	if !first.WorkTo.Equal(start.Add(18 * time.Hour)) {
		t.Fatalf("unexpected WorkTo: got=%v want=%v", first.WorkTo, start.Add(18*time.Hour))
	}
	if !first.BreakFrom.Equal(start.Add(13 * time.Hour)) {
		t.Fatalf("unexpected BreakFrom: got=%v want=%v", first.BreakFrom, start.Add(13*time.Hour))
	}
	if !first.BreakTo.Equal(start.Add(14 * time.Hour)) {
		t.Fatalf("unexpected BreakTo: got=%v want=%v", first.BreakTo, start.Add(14*time.Hour))
	}

	weekendFound := false
	weekdayFound := false
	for _, item := range items {
		if item.IsWeekEnd {
			weekendFound = true
		} else {
			weekdayFound = true
		}
	}
	if !weekendFound || !weekdayFound {
		t.Fatalf("expected both weekend and weekday entries, got weekend=%v weekday=%v", weekendFound, weekdayFound)
	}
}
