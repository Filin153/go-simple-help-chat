package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type departmentRepoMock struct {
	createFn  func(ctx context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error)
	getAllFn  func(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error)
	getByIDFn func(ctx context.Context, id int, tx pgx.Tx) (domain.Department, error)
	updateFn  func(ctx context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error
}

func (d *departmentRepoMock) Create(ctx context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error) {
	return d.createFn(ctx, createDepartment, tx)
}

func (d *departmentRepoMock) GetAll(ctx context.Context, page, limit int, tx pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error) {
	return d.getAllFn(ctx, page, limit, tx)
}

func (d *departmentRepoMock) GetByID(ctx context.Context, id int, tx pgx.Tx) (domain.Department, error) {
	return d.getByIDFn(ctx, id, tx)
}

func (d *departmentRepoMock) Update(ctx context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error {
	return d.updateFn(ctx, id, updateDepartment, tx)
}

type departmentScheduleUseCaseMock struct {
	generateBaseScheduleFn func(ctx context.Context, departmentID int) ([]domain.CreateSchedule, error)
}

func (d *departmentScheduleUseCaseMock) GenerateBaseSchedule(ctx context.Context, departmentID int) ([]domain.CreateSchedule, error) {
	return d.generateBaseScheduleFn(ctx, departmentID)
}

type departmentScheduleRepoMock struct {
	createFn    func(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error
	getFromToFn func(ctx context.Context, departmentID int, from, to time.Time, tx pgx.Tx) ([]domain.Schedule, error)
}

func (d *departmentScheduleRepoMock) Create(ctx context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
	return d.createFn(ctx, createSchedule, tx)
}

func (d *departmentScheduleRepoMock) GetFromTo(ctx context.Context, departmentID int, from, to time.Time, tx pgx.Tx) ([]domain.Schedule, error) {
	return d.getFromToFn(ctx, departmentID, from, to, tx)
}

type departmentFixture struct {
	useCase         *DepartmentUseCase
	tx              *fakeTx
	mainRepo        *mainRepoMock
	departmentRepo  *departmentRepoMock
	scheduleUseCase *departmentScheduleUseCaseMock
	scheduleRepo    *departmentScheduleRepoMock
	department      domain.Department
	departments     []domain.DepartmentWithOnScheduleeDay
	schedules       []domain.Schedule
	baseSchedule    []domain.CreateSchedule
}

func newDepartmentFixture() *departmentFixture {
	txObj := &fakeTx{}
	department := domain.Department{
		ID:   7,
		Name: "Support",
	}
	departments := []domain.DepartmentWithOnScheduleeDay{
		{
			ID:        department.ID,
			Name:      department.Name,
			WorkFrom:  time.Date(2026, time.April, 19, 9, 0, 0, 0, time.UTC),
			WorkTo:    time.Date(2026, time.April, 19, 18, 0, 0, 0, time.UTC),
			IsWeekEnd: false,
		},
	}
	schedules := []domain.Schedule{
		{
			ID:           1,
			DepartmentID: department.ID,
			Name:         "Monday",
			WorkFrom:     time.Date(2026, time.April, 20, 9, 0, 0, 0, time.UTC),
			WorkTo:       time.Date(2026, time.April, 20, 18, 0, 0, 0, time.UTC),
		},
	}
	baseSchedule := []domain.CreateSchedule{
		{
			DepartmentID: department.ID,
			Name:         "Monday",
			WorkFrom:     time.Date(2026, time.April, 20, 9, 0, 0, 0, time.UTC),
			WorkTo:       time.Date(2026, time.April, 20, 18, 0, 0, 0, time.UTC),
			BreakFrom:    time.Date(2026, time.April, 20, 13, 0, 0, 0, time.UTC),
			BreakTo:      time.Date(2026, time.April, 20, 14, 0, 0, 0, time.UTC),
		},
		{
			DepartmentID: department.ID,
			Name:         "Tuesday",
			WorkFrom:     time.Date(2026, time.April, 21, 9, 0, 0, 0, time.UTC),
			WorkTo:       time.Date(2026, time.April, 21, 18, 0, 0, 0, time.UTC),
			BreakFrom:    time.Date(2026, time.April, 21, 13, 0, 0, 0, time.UTC),
			BreakTo:      time.Date(2026, time.April, 21, 14, 0, 0, 0, time.UTC),
		},
	}

	mainRepo := &mainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return txObj, nil
		},
	}
	departmentRepo := &departmentRepoMock{
		createFn: func(_ context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error) {
			if tx != txObj {
				return 0, errors.New("expected tx")
			}
			if createDepartment.Name == "" {
				return 0, errors.New("empty department name")
			}
			return department.ID, nil
		},
		getAllFn: func(_ context.Context, page, limit int, tx pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error) {
			if tx != nil {
				return nil, errors.New("expected nil tx")
			}
			if page != 2 || limit != 10 {
				return nil, errors.New("unexpected pagination")
			}
			return departments, nil
		},
		getByIDFn: func(_ context.Context, id int, tx pgx.Tx) (domain.Department, error) {
			if tx != nil {
				return domain.Department{}, errors.New("expected nil tx")
			}
			if id != department.ID {
				return domain.Department{}, errors.New("unexpected department id")
			}
			return department, nil
		},
		updateFn: func(_ context.Context, id int, updateDepartment domain.UpdateDepartment, tx pgx.Tx) error {
			if tx != nil {
				return errors.New("expected nil tx")
			}
			if id != department.ID {
				return errors.New("unexpected department id")
			}
			if updateDepartment.Name == "" {
				return errors.New("empty update name")
			}
			return nil
		},
	}
	scheduleUseCase := &departmentScheduleUseCaseMock{
		generateBaseScheduleFn: func(_ context.Context, departmentID int) ([]domain.CreateSchedule, error) {
			if departmentID != department.ID {
				return nil, errors.New("unexpected department id")
			}
			return baseSchedule, nil
		},
	}
	scheduleRepo := &departmentScheduleRepoMock{
		createFn: func(_ context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
			if tx != txObj {
				return errors.New("expected tx")
			}
			if createSchedule == nil {
				return errors.New("expected schedule")
			}
			return nil
		},
		getFromToFn: func(_ context.Context, departmentID int, from, to time.Time, tx pgx.Tx) ([]domain.Schedule, error) {
			if tx != nil {
				return nil, errors.New("expected nil tx")
			}
			if departmentID != department.ID {
				return nil, errors.New("unexpected department id")
			}
			if !to.After(from) {
				return nil, errors.New("expected to after from")
			}
			return schedules, nil
		},
	}

	useCase := NewDepartmentUseCase(mainRepo, departmentRepo, scheduleUseCase, scheduleRepo)

	return &departmentFixture{
		useCase:         useCase,
		tx:              txObj,
		mainRepo:        mainRepo,
		departmentRepo:  departmentRepo,
		scheduleUseCase: scheduleUseCase,
		scheduleRepo:    scheduleRepo,
		department:      department,
		departments:     departments,
		schedules:       schedules,
		baseSchedule:    baseSchedule,
	}
}

func Test_NewDepartmentUseCase(t *testing.T) {
	f := newDepartmentFixture()

	if f.useCase == nil {
		t.Fatal("NewDepartmentUseCase returned nil")
	}
	if f.useCase.mainRepo != f.mainRepo {
		t.Fatal("unexpected mainRepo")
	}
	if f.useCase.departmentRepo != f.departmentRepo {
		t.Fatal("unexpected departmentRepo")
	}
	if f.useCase.scheduleUseCase != f.scheduleUseCase {
		t.Fatal("unexpected scheduleUseCase")
	}
	if f.useCase.scheduleRepo != f.scheduleRepo {
		t.Fatal("unexpected scheduleRepo")
	}
}

func Test_DepartmentUseCase_Create_CreateSessionError(t *testing.T) {
	f := newDepartmentFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
}

func Test_DepartmentUseCase_Create_DepartmentCreateError(t *testing.T) {
	f := newDepartmentFixture()
	wantErr := errors.New("create department error")
	f.departmentRepo.createFn = func(_ context.Context, createDepartment domain.CreateDepartment, tx pgx.Tx) (int, error) {
		if tx != f.tx {
			t.Fatal("expected tx in Create")
		}
		if createDepartment.Name != "Support" {
			t.Fatalf("unexpected department name: got=%q want=%q", createDepartment.Name, "Support")
		}
		return 0, wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create department error, got=%v", err)
	}
}

func Test_DepartmentUseCase_Create_GenerateBaseScheduleError(t *testing.T) {
	f := newDepartmentFixture()
	wantErr := errors.New("generate schedule error")
	f.scheduleUseCase.generateBaseScheduleFn = func(_ context.Context, departmentID int) ([]domain.CreateSchedule, error) {
		if departmentID != f.department.ID {
			t.Fatalf("unexpected department id: got=%d want=%d", departmentID, f.department.ID)
		}
		return nil, wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected generate base schedule error, got=%v", err)
	}
}

func Test_DepartmentUseCase_Create_ScheduleCreateError(t *testing.T) {
	f := newDepartmentFixture()
	wantErr := errors.New("create schedule error")
	callCount := 0
	f.scheduleRepo.createFn = func(_ context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
		callCount++
		if tx != f.tx {
			t.Fatal("expected tx in schedule Create")
		}
		if createSchedule.DepartmentID != f.department.ID {
			t.Fatalf("unexpected department id: got=%d want=%d", createSchedule.DepartmentID, f.department.ID)
		}
		return wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create schedule error, got=%v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected one create schedule call before failure, got=%d", callCount)
	}
}

func Test_DepartmentUseCase_Create_CommitError(t *testing.T) {
	f := newDepartmentFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_DepartmentUseCase_Create_OK(t *testing.T) {
	f := newDepartmentFixture()
	createdNames := make([]string, 0, len(f.baseSchedule))
	f.scheduleRepo.createFn = func(_ context.Context, createSchedule *domain.CreateSchedule, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("expected tx in schedule Create")
		}
		createdNames = append(createdNames, createSchedule.Name)
		return nil
	}

	err := f.useCase.Create(context.Background(), domain.CreateDepartment{Name: "Support"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if len(createdNames) != len(f.baseSchedule) {
		t.Fatalf("unexpected number of schedules created: got=%d want=%d", len(createdNames), len(f.baseSchedule))
	}
	if createdNames[0] != f.baseSchedule[0].Name || createdNames[1] != f.baseSchedule[1].Name {
		t.Fatalf("unexpected created schedule names: got=%v", createdNames)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit to be called once, got=%d", f.tx.commitCalls)
	}
}

func Test_DepartmentUseCase_GetAll_LimitError(t *testing.T) {
	f := newDepartmentFixture()

	departments, err := f.useCase.GetAll(context.Background(), 1, 101)
	if !errors.Is(err, domain.ErrLimitIsBiggerThen100) {
		t.Fatalf("expected limit error, got=%v", err)
	}
	if departments != nil {
		t.Fatalf("expected nil departments on error, got=%v", departments)
	}
}

func Test_DepartmentUseCase_GetAll_OK(t *testing.T) {
	f := newDepartmentFixture()

	departments, err := f.useCase.GetAll(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(departments) != 1 {
		t.Fatalf("unexpected departments len: got=%d want=1", len(departments))
	}
	if departments[0].ID != f.departments[0].ID {
		t.Fatalf("unexpected department id: got=%d want=%d", departments[0].ID, f.departments[0].ID)
	}
}

func Test_DepartmentUseCase_GetByID(t *testing.T) {
	f := newDepartmentFixture()

	department, err := f.useCase.GetByID(context.Background(), f.department.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if department.ID != f.department.ID {
		t.Fatalf("unexpected department id: got=%d want=%d", department.ID, f.department.ID)
	}
}

func Test_DepartmentUseCase_GetSheduleById_DurationError(t *testing.T) {
	f := newDepartmentFixture()
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(oneMonthDuration + time.Second)

	schedules, err := f.useCase.GetSheduleById(context.Background(), f.department.ID, from, to)
	if !errors.Is(err, domain.ErrDurationFromTo) {
		t.Fatalf("expected duration error, got=%v", err)
	}
	if schedules != nil {
		t.Fatalf("expected nil schedules on error, got=%v", schedules)
	}
}

func Test_DepartmentUseCase_GetSheduleById_OK(t *testing.T) {
	f := newDepartmentFixture()
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(oneMonthDuration)

	schedules, err := f.useCase.GetSheduleById(context.Background(), f.department.ID, from, to)
	if err != nil {
		t.Fatalf("GetSheduleById returned error: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("unexpected schedules len: got=%d want=1", len(schedules))
	}
	if schedules[0].DepartmentID != f.schedules[0].DepartmentID {
		t.Fatalf("unexpected department id: got=%d want=%d", schedules[0].DepartmentID, f.schedules[0].DepartmentID)
	}
}

func Test_DepartmentUseCase_Update(t *testing.T) {
	f := newDepartmentFixture()

	err := f.useCase.Update(context.Background(), f.department.ID, domain.UpdateDepartment{Name: "New Support"})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}
