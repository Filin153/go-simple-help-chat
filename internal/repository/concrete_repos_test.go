package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"shc/domain"

	"github.com/jackc/pgx/v5/pgconn"
	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func newMockRepository(t *testing.T) (pgxmock.PgxPoolIface, *Repository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	t.Cleanup(func() {
		mock.Close()
	})

	return mock, &Repository{
		db:      mock,
		beginTx: mock.BeginTx,
	}
}

func expectMock(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations were not met: %v", err)
	}
}

func Test_ConstructorsAndGetOffset(t *testing.T) {
	repo := &Repository{}
	if NewDepartmentRepo(repo).repo != repo {
		t.Fatal("unexpected department repo")
	}
	if NewMsgRepo(repo).repo != repo {
		t.Fatal("unexpected msg repo")
	}
	if NewRefreshTokenRepo(repo).repo != repo {
		t.Fatal("unexpected refresh token repo")
	}
	if NewScheduleRepo(repo).repo != repo {
		t.Fatal("unexpected schedule repo")
	}
	if NewTicketRepo(repo).repo != repo {
		t.Fatal("unexpected ticket repo")
	}
	if NewUserRepo(repo).repo != repo {
		t.Fatal("unexpected user repo")
	}
	if got := getOffset(3, 10); got != 20 {
		t.Fatalf("unexpected offset: got=%d want=20", got)
	}
}

func Test_DepartmentRepo(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)

		mock.ExpectQuery(`INSERT INTO "departments"`).
			WithArgs("Support").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(7))

		id, err := repo.Create(context.Background(), domain.CreateDepartment{Name: "Support"}, nil)
		if err != nil || id != 7 {
			t.Fatalf("unexpected create result: id=%d err=%v", id, err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)
		wantErr := errors.New("query error")

		mock.ExpectQuery(`INSERT INTO "departments"`).
			WithArgs("Support").
			WillReturnError(wantErr)

		_, err := repo.Create(context.Background(), domain.CreateDepartment{Name: "Support"}, nil)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateScanError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)

		mock.ExpectQuery(`INSERT INTO "departments"`).
			WithArgs("Support").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("bad"))

		if _, err := repo.Create(context.Background(), domain.CreateDepartment{Name: "Support"}, nil); err == nil {
			t.Fatal("expected scan error")
		}
		expectMock(t, mock)
	})

	t.Run("GetAll", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)
		now := time.Now()

		mock.ExpectQuery(`FROM "departments" AS d`).
			WithArgs(10, 10).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "work_from", "work_to", "is_week_end"}).
				AddRow(1, "Support", now, now.Add(time.Hour), false))

		items, err := repo.GetAll(context.Background(), 2, 10, nil)
		if err != nil || len(items) != 1 || items[0].ID != 1 {
			t.Fatalf("unexpected get all result: items=%v err=%v", items, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetAllQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)
		wantErr := errors.New("query error")

		mock.ExpectQuery(`FROM "departments" AS d`).
			WithArgs(10, 0).
			WillReturnError(wantErr)

		_, err := repo.GetAll(context.Background(), 1, 10, nil)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetAllCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)

		mock.ExpectQuery(`FROM "departments" AS d`).
			WithArgs(10, 0).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetAll(context.Background(), 1, 10, nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	t.Run("GetByID", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)

		mock.ExpectQuery(`SELECT "id", "name" FROM "departments"`).
			WithArgs(1).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).AddRow(1, "Support"))

		item, err := repo.GetByID(context.Background(), 1, nil)
		if err != nil || item.ID != 1 {
			t.Fatalf("unexpected get by id result: item=%v err=%v", item, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetByIDQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)
		wantErr := errors.New("query error")

		mock.ExpectQuery(`SELECT "id", "name" FROM "departments"`).
			WithArgs(1).
			WillReturnError(wantErr)

		if _, err := repo.GetByID(context.Background(), 1, nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetByIDCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewDepartmentRepo(baseRepo)

		mock.ExpectQuery(`SELECT "id", "name" FROM "departments"`).
			WithArgs(1).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetByID(context.Background(), 1, nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name    string
			result  pgconn.CommandTag
			err     error
			wantErr error
		}{
			{name: "success", result: pgxmock.NewResult("UPDATE", 1)},
			{name: "exec error", err: errors.New("exec error"), wantErr: errors.New("exec error")},
			{name: "unknown object", result: pgxmock.NewResult("UPDATE", 0), wantErr: domain.ErrUnknownObject},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mock, baseRepo := newMockRepository(t)
				repo := NewDepartmentRepo(baseRepo)

				exp := mock.ExpectExec(`UPDATE "departments" SET "name"=\$2 WHERE "id"=\$1;`).
					WithArgs(1, "Support")
				if tt.err != nil {
					exp.WillReturnError(tt.err)
				} else {
					exp.WillReturnResult(tt.result)
				}

				err := repo.Update(context.Background(), 1, domain.UpdateDepartment{Name: "Support"}, nil)
				if tt.wantErr == nil {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got=%v", tt.wantErr, err)
				}
				expectMock(t, mock)
			})
		}
	})
}

func Test_UserRepo(t *testing.T) {
	userRows := func() *pgxmock.Rows {
		return pgxmock.NewRows([]string{"uuid", "login", "password", "role"}).
			AddRow("uuid-1", "login", "hash", domain.UserRoleManager)
	}

	t.Run("GetAll", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewUserRepo(baseRepo)

		mock.ExpectQuery(`SELECT "uuid", "login", "password", "role"`).
			WithArgs(10, 10).
			WillReturnRows(userRows())

		users, err := repo.GetAll(context.Background(), 2, 10, nil)
		if err != nil || len(users) != 1 || users[0].UUID != "uuid-1" {
			t.Fatalf("unexpected get all result: users=%v err=%v", users, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetAllQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewUserRepo(baseRepo)
		wantErr := errors.New("query error")

		mock.ExpectQuery(`SELECT "uuid", "login", "password", "role"`).
			WithArgs(10, 0).
			WillReturnError(wantErr)

		if _, err := repo.GetAll(context.Background(), 1, 10, nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetAllCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewUserRepo(baseRepo)

		mock.ExpectQuery(`SELECT "uuid", "login", "password", "role"`).
			WithArgs(10, 0).
			WillReturnRows(pgxmock.NewRows([]string{"uuid"}).AddRow("uuid-1"))

		if _, err := repo.GetAll(context.Background(), 1, 10, nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	for _, tc := range []struct {
		name  string
		call  func(repo *UserRepo) (*domain.User, error)
		query string
		arg   string
	}{
		{name: "GetByUUID", call: func(repo *UserRepo) (*domain.User, error) { return repo.GetByUUID(context.Background(), "uuid-1", nil) }, query: `WHERE "uuid"=\$1;`, arg: "uuid-1"},
		{name: "GetByLogin", call: func(repo *UserRepo) (*domain.User, error) { return repo.GetByLogin(context.Background(), "login", nil) }, query: `WHERE "login"=\$1;`, arg: "login"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewUserRepo(baseRepo)
			mock.ExpectQuery(tc.query).WithArgs(tc.arg).WillReturnRows(userRows())

			user, err := tc.call(repo)
			if err != nil || user.UUID != "uuid-1" {
				t.Fatalf("unexpected user result: user=%v err=%v", user, err)
			}
			expectMock(t, mock)
		})

		t.Run(tc.name+"QueryError", func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewUserRepo(baseRepo)
			wantErr := errors.New("query error")
			mock.ExpectQuery(tc.query).WithArgs(tc.arg).WillReturnError(wantErr)

			if _, err := tc.call(repo); !errors.Is(err, wantErr) {
				t.Fatalf("expected query error, got=%v", err)
			}
			expectMock(t, mock)
		})

		t.Run(tc.name+"CollectError", func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewUserRepo(baseRepo)
			mock.ExpectQuery(tc.query).WithArgs(tc.arg).WillReturnRows(pgxmock.NewRows([]string{"uuid"}).AddRow("uuid-1"))

			if _, err := tc.call(repo); err == nil {
				t.Fatal("expected collect error")
			}
			expectMock(t, mock)
		})
	}

	for _, tc := range []struct {
		name    string
		call    func(repo *UserRepo) error
		query   string
		args    []any
		result  pgconn.CommandTag
		err     error
		wantErr error
	}{
		{
			name: "CreateSuccess",
			call: func(repo *UserRepo) error {
				return repo.Create(context.Background(), domain.CreateUser{Login: "login", Password: "hash", Role: domain.UserRoleManager}, nil)
			},
			query:  `INSERT INTO "users"`,
			args:   []any{"login", "hash", domain.UserRoleManager},
			result: pgxmock.NewResult("INSERT", 1),
		},
		{
			name: "CreateError",
			call: func(repo *UserRepo) error {
				return repo.Create(context.Background(), domain.CreateUser{Login: "login", Password: "hash", Role: domain.UserRoleManager}, nil)
			},
			query:   `INSERT INTO "users"`,
			args:    []any{"login", "hash", domain.UserRoleManager},
			err:     errors.New("exec error"),
			wantErr: errors.New("exec error"),
		},
		{
			name: "UpdateByUUIDSuccess",
			call: func(repo *UserRepo) error {
				return repo.UpdateByUUID(context.Background(), "uuid-1", domain.UpdateUser{CreateUser: domain.CreateUser{Login: "login", Password: "hash", Role: domain.UserRoleManager}}, nil)
			},
			query:  `UPDATE "users" SET "login"=\$2, "password"=\$3, "role"=\$4 WHERE "uuid"=\$1;`,
			args:   []any{"uuid-1", "login", "hash", domain.UserRoleManager},
			result: pgxmock.NewResult("UPDATE", 1),
		},
		{
			name: "UpdateByUUIDExecError",
			call: func(repo *UserRepo) error {
				return repo.UpdateByUUID(context.Background(), "uuid-1", domain.UpdateUser{CreateUser: domain.CreateUser{Login: "login", Password: "hash", Role: domain.UserRoleManager}}, nil)
			},
			query:   `UPDATE "users" SET "login"=\$2, "password"=\$3, "role"=\$4 WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1", "login", "hash", domain.UserRoleManager},
			err:     errors.New("exec error"),
			wantErr: errors.New("exec error"),
		},
		{
			name: "UpdateByUUIDUnknownObject",
			call: func(repo *UserRepo) error {
				return repo.UpdateByUUID(context.Background(), "uuid-1", domain.UpdateUser{CreateUser: domain.CreateUser{Login: "login", Password: "hash", Role: domain.UserRoleManager}}, nil)
			},
			query:   `UPDATE "users" SET "login"=\$2, "password"=\$3, "role"=\$4 WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1", "login", "hash", domain.UserRoleManager},
			result:  pgxmock.NewResult("UPDATE", 0),
			wantErr: domain.ErrUnknownObject,
		},
		{
			name:   "DeleteByUUIDSuccess",
			call:   func(repo *UserRepo) error { return repo.DeleteByUUID(context.Background(), "uuid-1", nil) },
			query:  `DELETE FROM "users" WHERE "uuid"=\$1;`,
			args:   []any{"uuid-1"},
			result: pgxmock.NewResult("DELETE", 1),
		},
		{
			name:    "DeleteByUUIDExecError",
			call:    func(repo *UserRepo) error { return repo.DeleteByUUID(context.Background(), "uuid-1", nil) },
			query:   `DELETE FROM "users" WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1"},
			err:     errors.New("exec error"),
			wantErr: errors.New("exec error"),
		},
		{
			name:    "DeleteByUUIDUnknownObject",
			call:    func(repo *UserRepo) error { return repo.DeleteByUUID(context.Background(), "uuid-1", nil) },
			query:   `DELETE FROM "users" WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1"},
			result:  pgxmock.NewResult("DELETE", 0),
			wantErr: domain.ErrUnknownObject,
		},
		{
			name: "UpdateUserPasswordByUUIDSuccess",
			call: func(repo *UserRepo) error {
				return repo.UpdateUserPasswordByUUID(context.Background(), "uuid-1", "hash", nil)
			},
			query:  `UPDATE "users" SET "password"=\$2 WHERE "uuid"=\$1;`,
			args:   []any{"uuid-1", "hash"},
			result: pgxmock.NewResult("UPDATE", 1),
		},
		{
			name: "UpdateUserPasswordByUUIDExecError",
			call: func(repo *UserRepo) error {
				return repo.UpdateUserPasswordByUUID(context.Background(), "uuid-1", "hash", nil)
			},
			query:   `UPDATE "users" SET "password"=\$2 WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1", "hash"},
			err:     errors.New("exec error"),
			wantErr: errors.New("exec error"),
		},
		{
			name: "UpdateUserPasswordByUUIDUnknownObject",
			call: func(repo *UserRepo) error {
				return repo.UpdateUserPasswordByUUID(context.Background(), "uuid-1", "hash", nil)
			},
			query:   `UPDATE "users" SET "password"=\$2 WHERE "uuid"=\$1;`,
			args:    []any{"uuid-1", "hash"},
			result:  pgxmock.NewResult("UPDATE", 0),
			wantErr: domain.ErrUnknownObject,
		},
		{
			name: "CreateClientSuccess",
			call: func(repo *UserRepo) error {
				return repo.CreateClient(context.Background(), domain.CreateClient{UserUUID: "uuid-1", Info: map[any]any{"x": "y"}}, nil)
			},
			query:  `INSERT INTO "clients"`,
			args:   []any{"uuid-1", map[any]any{"x": "y"}},
			result: pgxmock.NewResult("INSERT", 1),
		},
		{
			name: "CreateClientError",
			call: func(repo *UserRepo) error {
				return repo.CreateClient(context.Background(), domain.CreateClient{UserUUID: "uuid-1", Info: map[any]any{"x": "y"}}, nil)
			},
			query:   `INSERT INTO "clients"`,
			args:    []any{"uuid-1", map[any]any{"x": "y"}},
			err:     errors.New("exec error"),
			wantErr: errors.New("exec error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewUserRepo(baseRepo)

			exp := mock.ExpectExec(tc.query).WithArgs(tc.args...)
			if tc.err != nil {
				exp.WillReturnError(tc.err)
			} else {
				exp.WillReturnResult(tc.result)
			}

			err := tc.call(repo)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
				t.Fatalf("expected error %v, got=%v", tc.wantErr, err)
			}
			expectMock(t, mock)
		})
	}
}

func Test_RefreshTokenRepo(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewRefreshTokenRepo(baseRepo)
		mock.ExpectQuery(`FROM "refresh_tokens" WHERE "jti"`).
			WithArgs("jti-1").
			WillReturnRows(pgxmock.NewRows([]string{"jti", "user_uuid"}).AddRow("jti-1", "uuid-1"))

		token, err := repo.Get(context.Background(), "jti-1", nil)
		if err != nil || token.JTI != "jti-1" {
			t.Fatalf("unexpected get result: token=%v err=%v", token, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewRefreshTokenRepo(baseRepo)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`FROM "refresh_tokens" WHERE "jti"`).WithArgs("jti-1").WillReturnError(wantErr)

		if _, err := repo.Get(context.Background(), "jti-1", nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewRefreshTokenRepo(baseRepo)
		mock.ExpectQuery(`FROM "refresh_tokens" WHERE "jti"`).WithArgs("jti-1").WillReturnRows(pgxmock.NewRows([]string{"jti"}).AddRow("jti-1"))

		if _, err := repo.Get(context.Background(), "jti-1", nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	for _, tc := range []struct {
		name    string
		call    func(repo *RefreshTokenRepo) error
		query   string
		args    []any
		result  pgconn.CommandTag
		err     error
		wantErr error
	}{
		{name: "CreateSuccess", call: func(repo *RefreshTokenRepo) error { return repo.Create(context.Background(), "jti", "uuid", nil) }, query: `INSERT INTO "refresh_tokens"`, args: []any{"jti", "uuid"}, result: pgxmock.NewResult("INSERT", 1)},
		{name: "CreateError", call: func(repo *RefreshTokenRepo) error { return repo.Create(context.Background(), "jti", "uuid", nil) }, query: `INSERT INTO "refresh_tokens"`, args: []any{"jti", "uuid"}, err: errors.New("exec error"), wantErr: errors.New("exec error")},
		{name: "DeleteByUserUUIDSuccess", call: func(repo *RefreshTokenRepo) error { return repo.DeleteByUserUUID(context.Background(), "uuid", nil) }, query: `DELETE FROM "refresh_tokens" WHERE user_uuid = \$1;`, args: []any{"uuid"}, result: pgxmock.NewResult("DELETE", 1)},
		{name: "DeleteByUserUUIDError", call: func(repo *RefreshTokenRepo) error { return repo.DeleteByUserUUID(context.Background(), "uuid", nil) }, query: `DELETE FROM "refresh_tokens" WHERE user_uuid = \$1;`, args: []any{"uuid"}, err: errors.New("exec error"), wantErr: errors.New("exec error")},
		{name: "DeleteByJTISuccess", call: func(repo *RefreshTokenRepo) error { return repo.DeleteByJTI(context.Background(), "jti", nil) }, query: `DELETE FROM "refresh_tokens" WHERE jti = \$1;`, args: []any{"jti"}, result: pgxmock.NewResult("DELETE", 1)},
		{name: "DeleteByJTIError", call: func(repo *RefreshTokenRepo) error { return repo.DeleteByJTI(context.Background(), "jti", nil) }, query: `DELETE FROM "refresh_tokens" WHERE jti = \$1;`, args: []any{"jti"}, err: errors.New("exec error"), wantErr: errors.New("exec error")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewRefreshTokenRepo(baseRepo)
			exp := mock.ExpectExec(tc.query).WithArgs(tc.args...)
			if tc.err != nil {
				exp.WillReturnError(tc.err)
			} else {
				exp.WillReturnResult(tc.result)
			}
			err := tc.call(repo)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
				t.Fatalf("expected error %v, got=%v", tc.wantErr, err)
			}
			expectMock(t, mock)
		})
	}
}

func Test_ScheduleRepo(t *testing.T) {
	now := time.Now()

	t.Run("Create", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		item := &domain.CreateSchedule{DepartmentID: 1, Name: "Mon", WorkFrom: now, WorkTo: now.Add(time.Hour), BreakFrom: now.Add(2 * time.Hour), BreakTo: now.Add(3 * time.Hour)}

		mock.ExpectExec(`INSERT INTO`).
			WithArgs(item.DepartmentID, item.Name, item.WorkFrom, item.WorkTo, item.BreakFrom, item.BreakTo, item.IsWeekEnd).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		if err := repo.Create(context.Background(), item, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		item := &domain.CreateSchedule{DepartmentID: 1, Name: "Mon", WorkFrom: now, WorkTo: now.Add(time.Hour), BreakFrom: now.Add(2 * time.Hour), BreakTo: now.Add(3 * time.Hour)}
		wantErr := errors.New("exec error")

		mock.ExpectExec(`INSERT INTO`).
			WithArgs(item.DepartmentID, item.Name, item.WorkFrom, item.WorkTo, item.BreakFrom, item.BreakTo, item.IsWeekEnd).
			WillReturnError(wantErr)

		if err := repo.Create(context.Background(), item, nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected exec error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name    string
			item    *domain.UpdateSchedule
			query   string
			args    []any
			result  pgconn.CommandTag
			err     error
			wantErr error
		}{
			{name: "empty object", item: &domain.UpdateSchedule{ID: 1}, wantErr: domain.ErrEmptyObject},
			{name: "exec error", item: &domain.UpdateSchedule{ID: 1, Name: "Mon"}, query: regexp.QuoteMeta(`UPDATE "schedules" SET "name" = $1 WHERE "id" = $2;`), args: []any{"Mon", 1}, err: errors.New("exec error"), wantErr: errors.New("exec error")},
			{name: "rows affected", item: &domain.UpdateSchedule{ID: 1, Name: "Mon"}, query: regexp.QuoteMeta(`UPDATE "schedules" SET "name" = $1 WHERE "id" = $2;`), args: []any{"Mon", 1}, result: pgxmock.NewResult("UPDATE", 0), wantErr: domain.ErrZeroRowAffected},
			{name: "success", item: &domain.UpdateSchedule{ID: 1, Name: "Mon"}, query: regexp.QuoteMeta(`UPDATE "schedules" SET "name" = $1 WHERE "id" = $2;`), args: []any{"Mon", 1}, result: pgxmock.NewResult("UPDATE", 1)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mock, baseRepo := newMockRepository(t)
				repo := NewScheduleRepo(baseRepo)

				if tt.query != "" {
					exp := mock.ExpectExec(tt.query).WithArgs(tt.args...)
					if tt.err != nil {
						exp.WillReturnError(tt.err)
					} else {
						exp.WillReturnResult(tt.result)
					}
				}

				err := repo.Update(context.Background(), tt.item, nil)
				if tt.wantErr == nil {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got=%v", tt.wantErr, err)
				}
				expectMock(t, mock)
			})
		}
	})

	for _, tc := range []struct {
		name    string
		call    func(repo *ScheduleRepo) error
		query   string
		arg     any
		result  pgconn.CommandTag
		err     error
		wantErr error
	}{
		{name: "DeleteSuccess", call: func(repo *ScheduleRepo) error { return repo.Delete(context.Background(), 1, nil) }, query: `DELETE FROM "schedules" WHERE id=\$1;`, arg: 1, result: pgxmock.NewResult("DELETE", 1)},
		{name: "DeleteExecError", call: func(repo *ScheduleRepo) error { return repo.Delete(context.Background(), 1, nil) }, query: `DELETE FROM "schedules" WHERE id=\$1;`, arg: 1, err: errors.New("exec error"), wantErr: errors.New("exec error")},
		{name: "DeleteRowsAffected", call: func(repo *ScheduleRepo) error { return repo.Delete(context.Background(), 1, nil) }, query: `DELETE FROM "schedules" WHERE id=\$1;`, arg: 1, result: pgxmock.NewResult("DELETE", 0), wantErr: domain.ErrZeroRowAffected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock, baseRepo := newMockRepository(t)
			repo := NewScheduleRepo(baseRepo)
			exp := mock.ExpectExec(tc.query).WithArgs(tc.arg)
			if tc.err != nil {
				exp.WillReturnError(tc.err)
			} else {
				exp.WillReturnResult(tc.result)
			}
			err := tc.call(repo)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
				t.Fatalf("expected error %v, got=%v", tc.wantErr, err)
			}
			expectMock(t, mock)
		})
	}

	t.Run("ExistByDepartmentID", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		mock.ExpectQuery(`SELECT count\("id"\) FROM "schedules"`).WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

		ok, err := repo.ExistByDepartmentID(context.Background(), 1)
		if err != nil || !ok {
			t.Fatalf("unexpected exist result: ok=%v err=%v", ok, err)
		}
		expectMock(t, mock)
	})

	t.Run("ExistByDepartmentIDQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`SELECT count\("id"\) FROM "schedules"`).WithArgs(1).WillReturnError(wantErr)

		if _, err := repo.ExistByDepartmentID(context.Background(), 1); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetFromTo", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		from := now
		to := now.Add(time.Hour)
		mock.ExpectQuery(`FROM "schedules"`).WithArgs(1, from, to).WillReturnRows(
			pgxmock.NewRows([]string{"id", "department_id", "name", "work_from", "work_to", "break_from", "break_to", "is_week_end"}).
				AddRow(1, 1, "Mon", from, to, from.Add(15*time.Minute), from.Add(30*time.Minute), false),
		)

		items, err := repo.GetFromTo(context.Background(), 1, from, to, nil)
		if err != nil || len(items) != 1 || items[0].ID != 1 {
			t.Fatalf("unexpected get from to result: items=%v err=%v", items, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetFromToQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		from := now
		to := now.Add(time.Hour)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`FROM "schedules"`).WithArgs(1, from, to).WillReturnError(wantErr)

		if _, err := repo.GetFromTo(context.Background(), 1, from, to, nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetFromToCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewScheduleRepo(baseRepo)
		from := now
		to := now.Add(time.Hour)
		mock.ExpectQuery(`FROM "schedules"`).WithArgs(1, from, to).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetFromTo(context.Background(), 1, from, to, nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})
}

func Test_TicketRepo(t *testing.T) {
	t.Run("GetByID", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewTicketRepo(baseRepo)
		now := time.Now()
		mock.ExpectQuery(`FROM "tickets"`).WithArgs(1).WillReturnRows(
			pgxmock.NewRows([]string{"id", "department_id", "name", "manager_user_uuid", "client_user_uuid", "status", "create_at", "update_at"}).
				AddRow(1, 2, "Ticket", "manager", "client", domain.NewTicketStatus, now, now),
		)

		item, err := repo.GetByID(context.Background(), 1)
		if err != nil || item.ID != 1 {
			t.Fatalf("unexpected ticket result: item=%v err=%v", item, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetByIDQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewTicketRepo(baseRepo)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`FROM "tickets"`).WithArgs(1).WillReturnError(wantErr)

		if _, err := repo.GetByID(context.Background(), 1); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetByIDCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewTicketRepo(baseRepo)
		mock.ExpectQuery(`FROM "tickets"`).WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetByID(context.Background(), 1); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})
}

func Test_MsgRepo(t *testing.T) {
	now := time.Now()
	msg := domain.CreateMsg{TicketID: 1, EncryptedText: []byte("enc")}

	t.Run("CreateForClient", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`INSERT INTO "messages"`).
			WithArgs(1, domain.MsgFromTypeManager, []byte("enc"), domain.MsgStatusSent).
			WillReturnRows(pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeManager, []byte("enc"), 1, domain.MsgStatusSent, now))

		item, err := repo.CreateForClient(context.Background(), msg, nil)
		if err != nil || item.ID != 1 {
			t.Fatalf("unexpected create for client result: item=%v err=%v", item, err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateForManager", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`INSERT INTO "messages"`).
			WithArgs(1, domain.MsgFromTypeClient, []byte("enc"), domain.MsgStatusSent).
			WillReturnRows(pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeClient, []byte("enc"), 1, domain.MsgStatusSent, now))

		item, err := repo.CreateForManager(context.Background(), msg, nil)
		if err != nil || item.ID != 1 {
			t.Fatalf("unexpected create for manager result: item=%v err=%v", item, err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateHelperErrors", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		wantErr := errors.New("query error")

		mock.ExpectQuery(`insert`).WithArgs(1, domain.MsgFromTypeSystem, []byte("enc"), domain.MsgStatusSent).WillReturnError(wantErr)
		if _, err := repo.create(context.Background(), `insert`, domain.CreateMsg{TicketID: 1, EncryptedText: []byte("enc")}, domain.MsgFromTypeSystem, nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)

		mock, baseRepo = newMockRepository(t)
		repo = NewMsgRepo(baseRepo)
		mock.ExpectQuery(`insert`).WithArgs(1, domain.MsgFromTypeSystem, []byte("enc"), domain.MsgStatusSent).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
		if _, err := repo.create(context.Background(), `insert`, domain.CreateMsg{TicketID: 1, EncryptedText: []byte("enc")}, domain.MsgFromTypeSystem, nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	t.Run("CreateFile", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`INSERT INTO "message_files"`).WithArgs(1, "file.txt", "/tmp/file").WillReturnRows(
			pgxmock.NewRows([]string{"id", "msg_id", "file_name", "path"}).AddRow(1, 1, "file.txt", "/tmp/file"),
		)

		item, err := repo.CreateFile(context.Background(), 1, "file.txt", "/tmp/file", nil)
		if err != nil || item.ID != 1 {
			t.Fatalf("unexpected create file result: item=%v err=%v", item, err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateFileQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`INSERT INTO "message_files"`).WithArgs(1, "file.txt", "/tmp/file").WillReturnError(wantErr)

		if _, err := repo.CreateFile(context.Background(), 1, "file.txt", "/tmp/file", nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("CreateFileCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`INSERT INTO "message_files"`).WithArgs(1, "file.txt", "/tmp/file").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.CreateFile(context.Background(), 1, "file.txt", "/tmp/file", nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	t.Run("GetUnread", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`FROM "messages"`).
			WithArgs("uuid-1", domain.MsgStatusSent).
			WillReturnRows(pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeManager, []byte("enc"), 1, domain.MsgStatusSent, now))
		mock.ExpectQuery(`FROM "message_files"`).
			WithArgs([]int{1}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "msg_id", "file_name", "path"}).
				AddRow(1, 1, "a.txt", "/a.txt"))

		items, err := repo.GetUnread(context.Background(), "uuid-1", nil)
		if err != nil || len(items) != 1 || len(items[0].Files) != 1 {
			t.Fatalf("unexpected unread result: items=%v err=%v", items, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetUnreadQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		wantErr := errors.New("query error")
		mock.ExpectQuery(`FROM "messages"`).WithArgs("uuid-1", domain.MsgStatusSent).WillReturnError(wantErr)

		if _, err := repo.GetUnread(context.Background(), "uuid-1", nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetUnreadCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`FROM "messages"`).WithArgs("uuid-1", domain.MsgStatusSent).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetUnread(context.Background(), "uuid-1", nil); err == nil {
			t.Fatal("expected collect error")
		}
		expectMock(t, mock)
	})

	t.Run("GetUnreadEmpty", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`FROM "messages"`).WithArgs("uuid-1", domain.MsgStatusSent).WillReturnRows(
			pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}),
		)

		items, err := repo.GetUnread(context.Background(), "uuid-1", nil)
		if err != nil || len(items) != 0 {
			t.Fatalf("unexpected unread result: items=%v err=%v", items, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetUnreadFilesQueryError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		wantErr := errors.New("files query error")
		mock.ExpectQuery(`FROM "messages"`).
			WithArgs("uuid-1", domain.MsgStatusSent).
			WillReturnRows(pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeManager, []byte("enc"), 1, domain.MsgStatusSent, now))
		mock.ExpectQuery(`FROM "message_files"`).WithArgs([]int{1}).WillReturnError(wantErr)

		if _, err := repo.GetUnread(context.Background(), "uuid-1", nil); !errors.Is(err, wantErr) {
			t.Fatalf("expected files query error, got=%v", err)
		}
		expectMock(t, mock)
	})

	t.Run("GetUnreadFilesCollectError", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		mock.ExpectQuery(`FROM "messages"`).
			WithArgs("uuid-1", domain.MsgStatusSent).
			WillReturnRows(pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeManager, []byte("enc"), 1, domain.MsgStatusSent, now))
		mock.ExpectQuery(`FROM "message_files"`).WithArgs([]int{1}).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		if _, err := repo.GetUnread(context.Background(), "uuid-1", nil); err == nil {
			t.Fatal("expected files collect error")
		}
		expectMock(t, mock)
	})

	t.Run("MarkReadByID", func(t *testing.T) {
		tests := []struct {
			name    string
			result  pgconn.CommandTag
			err     error
			wantErr error
		}{
			{name: "success", result: pgxmock.NewResult("UPDATE", 1)},
			{name: "exec error", err: errors.New("exec error"), wantErr: errors.New("exec error")},
			{name: "unknown object", result: pgxmock.NewResult("UPDATE", 0), wantErr: domain.ErrUnknownObject},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mock, baseRepo := newMockRepository(t)
				repo := NewMsgRepo(baseRepo)
				exp := mock.ExpectExec(`UPDATE "messages"`).WithArgs("uuid-1", 1, domain.MsgStatusRead)
				if tt.err != nil {
					exp.WillReturnError(tt.err)
				} else {
					exp.WillReturnResult(tt.result)
				}

				err := repo.MarkReadByID(context.Background(), "uuid-1", 1, nil)
				if tt.wantErr == nil {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got=%v", tt.wantErr, err)
				}
				expectMock(t, mock)
			})
		}
	})

	t.Run("GetHistory", func(t *testing.T) {
		mock, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		from := now.Add(-time.Hour)
		to := now
		mock.ExpectQuery(`JOIN "tickets" AS t`).WithArgs("uuid-1", 1, from, to).WillReturnRows(
			pgxmock.NewRows([]string{"id", "from_type", "text", "ticket_id", "status", "create_at"}).
				AddRow(1, domain.MsgFromTypeManager, []byte("enc"), 1, domain.MsgStatusSent, now),
		)
		mock.ExpectQuery(`FROM "message_files"`).WithArgs([]int{1}).WillReturnRows(
			pgxmock.NewRows([]string{"id", "msg_id", "file_name", "path"}),
		)

		items, err := repo.GetHistory(context.Background(), "uuid-1", 1, from, to)
		if err != nil || len(items) != 1 {
			t.Fatalf("unexpected history result: items=%v err=%v", items, err)
		}
		expectMock(t, mock)
	})

	t.Run("GetFilesByMsgIDsEmpty", func(t *testing.T) {
		_, baseRepo := newMockRepository(t)
		repo := NewMsgRepo(baseRepo)
		items, err := repo.getFilesByMsgIDs(context.Background(), nil, nil)
		if err != nil || len(items) != 0 {
			t.Fatalf("unexpected files result: items=%v err=%v", items, err)
		}
	})

	t.Run("MsgIDs", func(t *testing.T) {
		got := msgIDs([]*domain.Msg{{ID: 1}, {ID: 2}})
		if len(got) != 2 || got[0] != 1 || got[1] != 2 {
			t.Fatalf("unexpected ids: %v", got)
		}
	})
}
