package usecase

import (
	"context"
	"errors"
	"testing"

	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type userMainRepoMock struct {
	createSessionFn func(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

func (m *userMainRepoMock) CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return m.createSessionFn(ctx, options)
}

type userCRUDRepoMock struct {
	getAllFn       func(ctx context.Context, tx *pgx.Tx) ([]domain.User, error)
	getByUUIDFn    func(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error)
	getByLoginFn   func(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
	createFn       func(ctx context.Context, user domain.CreateUser, tx *pgx.Tx) error
	updateByUUIDFn func(ctx context.Context, uuid string, user domain.UpdateUser, tx *pgx.Tx) error
	deleteByUUIDFn func(ctx context.Context, uuid string, tx *pgx.Tx) error
}

func (r *userCRUDRepoMock) GetAll(ctx context.Context, tx *pgx.Tx) ([]domain.User, error) {
	return r.getAllFn(ctx, tx)
}

func (r *userCRUDRepoMock) GetByUUID(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error) {
	return r.getByUUIDFn(ctx, uuid, tx)
}

func (r *userCRUDRepoMock) GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error) {
	return r.getByLoginFn(ctx, login, tx)
}

func (r *userCRUDRepoMock) Create(ctx context.Context, user domain.CreateUser, tx *pgx.Tx) error {
	return r.createFn(ctx, user, tx)
}

func (r *userCRUDRepoMock) UpdateByUUID(ctx context.Context, uuid string, user domain.UpdateUser, tx *pgx.Tx) error {
	return r.updateByUUIDFn(ctx, uuid, user, tx)
}

func (r *userCRUDRepoMock) DeleteByUUID(ctx context.Context, uuid string, tx *pgx.Tx) error {
	return r.deleteByUUIDFn(ctx, uuid, tx)
}

type userPswdServiceMock struct {
	createPasswordHashFn func(password string) (string, error)
}

func (p *userPswdServiceMock) CreatePasswordHash(password string) (string, error) {
	return p.createPasswordHashFn(password)
}

type userFixture struct {
	useCase     *UserUseCase
	mainRepo    *userMainRepoMock
	userRepo    *userCRUDRepoMock
	pswdService *userPswdServiceMock
	users       []domain.User
	user        *domain.User
}

func newUserFixture() *userFixture {
	user := &domain.User{
		UUID:     "user-uuid-1",
		Login:    "login",
		Password: "hashed-password",
		Role:     domain.UserRoleManager,
	}
	users := []domain.User{*user}

	mainRepo := &userMainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return nil, nil
		},
	}
	userRepo := &userCRUDRepoMock{
		getAllFn: func(_ context.Context, tx *pgx.Tx) ([]domain.User, error) {
			if tx != nil {
				return nil, errors.New("expected nil tx")
			}
			return users, nil
		},
		getByUUIDFn: func(_ context.Context, uuid string, tx *pgx.Tx) (*domain.User, error) {
			if tx != nil {
				return nil, errors.New("expected nil tx")
			}
			if uuid != user.UUID {
				return nil, errors.New("unexpected uuid")
			}
			return user, nil
		},
		getByLoginFn: func(_ context.Context, login string, tx *pgx.Tx) (*domain.User, error) {
			if tx != nil {
				return nil, errors.New("expected nil tx")
			}
			if login != user.Login {
				return nil, errors.New("unexpected login")
			}
			return user, nil
		},
		createFn: func(_ context.Context, _ domain.CreateUser, tx *pgx.Tx) error {
			if tx != nil {
				return errors.New("expected nil tx")
			}
			return nil
		},
		updateByUUIDFn: func(_ context.Context, _ string, _ domain.UpdateUser, tx *pgx.Tx) error {
			if tx != nil {
				return errors.New("expected nil tx")
			}
			return nil
		},
		deleteByUUIDFn: func(_ context.Context, _ string, tx *pgx.Tx) error {
			if tx != nil {
				return errors.New("expected nil tx")
			}
			return nil
		},
	}
	pswdService := &userPswdServiceMock{
		createPasswordHashFn: func(password string) (string, error) {
			if password == "" {
				return "", errors.New("password is empty")
			}
			return "hashed-password", nil
		},
	}

	useCase := NewUserUseCase(mainRepo, userRepo, pswdService)

	return &userFixture{
		useCase:     useCase,
		mainRepo:    mainRepo,
		userRepo:    userRepo,
		pswdService: pswdService,
		users:       users,
		user:        user,
	}
}

func Test_NewUserUseCase(t *testing.T) {
	f := newUserFixture()
	if f.useCase == nil {
		t.Fatal("NewUserUseCase returned nil")
	}
}

func Test_UserUseCase_GetAll(t *testing.T) {
	f := newUserFixture()

	users, err := f.useCase.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("unexpected users len: got=%d want=1", len(users))
	}
	if users[0].UUID != f.user.UUID {
		t.Fatalf("unexpected user uuid: got=%q want=%q", users[0].UUID, f.user.UUID)
	}
}

func Test_UserUseCase_GetByUUID(t *testing.T) {
	f := newUserFixture()

	user, err := f.useCase.GetByUUID(context.Background(), f.user.UUID)
	if err != nil {
		t.Fatalf("GetByUUID returned error: %v", err)
	}
	if user.UUID != f.user.UUID {
		t.Fatalf("unexpected user uuid: got=%q want=%q", user.UUID, f.user.UUID)
	}
}

func Test_UserUseCase_GetByLogin(t *testing.T) {
	f := newUserFixture()

	user, err := f.useCase.GetByLogin(context.Background(), f.user.Login)
	if err != nil {
		t.Fatalf("GetByLogin returned error: %v", err)
	}
	if user.Login != f.user.Login {
		t.Fatalf("unexpected user login: got=%q want=%q", user.Login, f.user.Login)
	}
}

func Test_UserUseCase_Create_HashError(t *testing.T) {
	f := newUserFixture()
	wantErr := errors.New("hash error")
	f.pswdService.createPasswordHashFn = func(_ string) (string, error) {
		return "", wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected hash error, got=%v", err)
	}
}

func Test_UserUseCase_Create_RepoError(t *testing.T) {
	f := newUserFixture()
	wantErr := errors.New("repo error")
	f.userRepo.createFn = func(_ context.Context, _ domain.CreateUser, _ *pgx.Tx) error {
		return wantErr
	}

	err := f.useCase.Create(context.Background(), domain.CreateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got=%v", err)
	}
}

func Test_UserUseCase_Create_OK(t *testing.T) {
	f := newUserFixture()
	called := false
	f.userRepo.createFn = func(_ context.Context, user domain.CreateUser, tx *pgx.Tx) error {
		called = true
		if tx != nil {
			t.Fatal("expected nil tx")
		}
		if user.Password != "hashed-password" {
			t.Fatalf("expected hashed password, got=%q", user.Password)
		}
		return nil
	}

	err := f.useCase.Create(context.Background(), domain.CreateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if !called {
		t.Fatal("expected repo Create to be called")
	}
}

func Test_UserUseCase_Create_ShortPassword(t *testing.T) {
	f := newUserFixture()
	hashCalled := false
	f.pswdService.createPasswordHashFn = func(_ string) (string, error) {
		hashCalled = true
		return "", nil
	}
	repoCalled := false
	f.userRepo.createFn = func(_ context.Context, _ domain.CreateUser, _ *pgx.Tx) error {
		repoCalled = true
		return nil
	}

	err := f.useCase.Create(context.Background(), domain.CreateUser{Login: "login", Password: "12345", Role: domain.UserRoleClient})
	if !errors.Is(err, domain.ErrShortPassword) {
		t.Fatalf("expected short password error, got=%v", err)
	}
	if hashCalled {
		t.Fatal("password hash must not be called for short password")
	}
	if repoCalled {
		t.Fatal("repo Create must not be called for short password")
	}
}

func Test_UserUseCase_UpdateByUUID_HashError(t *testing.T) {
	f := newUserFixture()
	wantErr := errors.New("hash error")
	f.pswdService.createPasswordHashFn = func(_ string) (string, error) {
		return "", wantErr
	}

	err := f.useCase.UpdateByUUID(context.Background(), f.user.UUID, domain.UpdateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected hash error, got=%v", err)
	}
}

func Test_UserUseCase_UpdateByUUID_RepoError(t *testing.T) {
	f := newUserFixture()
	wantErr := errors.New("repo error")
	f.userRepo.updateByUUIDFn = func(_ context.Context, _ string, _ domain.UpdateUser, _ *pgx.Tx) error {
		return wantErr
	}

	err := f.useCase.UpdateByUUID(context.Background(), f.user.UUID, domain.UpdateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got=%v", err)
	}
}

func Test_UserUseCase_UpdateByUUID_OK(t *testing.T) {
	f := newUserFixture()
	called := false
	f.userRepo.updateByUUIDFn = func(_ context.Context, uuid string, user domain.UpdateUser, tx *pgx.Tx) error {
		called = true
		if tx != nil {
			t.Fatal("expected nil tx")
		}
		if uuid != f.user.UUID {
			t.Fatalf("unexpected uuid: got=%q want=%q", uuid, f.user.UUID)
		}
		if user.Password != "hashed-password" {
			t.Fatalf("expected hashed password, got=%q", user.Password)
		}
		return nil
	}

	err := f.useCase.UpdateByUUID(context.Background(), f.user.UUID, domain.UpdateUser{Login: "login", Password: "password", Role: domain.UserRoleClient})
	if err != nil {
		t.Fatalf("UpdateByUUID returned error: %v", err)
	}
	if !called {
		t.Fatal("expected repo UpdateByUUID to be called")
	}
}

func Test_UserUseCase_UpdateByUUID_WithoutPassword(t *testing.T) {
	f := newUserFixture()
	hashCalled := false
	f.pswdService.createPasswordHashFn = func(_ string) (string, error) {
		hashCalled = true
		return "", nil
	}
	f.userRepo.updateByUUIDFn = func(_ context.Context, _ string, user domain.UpdateUser, tx *pgx.Tx) error {
		if tx != nil {
			t.Fatal("expected nil tx")
		}
		if user.Password != "" {
			t.Fatalf("expected empty password, got=%q", user.Password)
		}
		return nil
	}

	err := f.useCase.UpdateByUUID(context.Background(), f.user.UUID, domain.UpdateUser{Login: "login", Password: "", Role: domain.UserRoleClient})
	if err != nil {
		t.Fatalf("UpdateByUUID returned error: %v", err)
	}
	if hashCalled {
		t.Fatal("password hash must not be called for empty password")
	}
}

func Test_UserUseCase_UpdateByUUID_ShortPassword(t *testing.T) {
	f := newUserFixture()
	hashCalled := false
	f.pswdService.createPasswordHashFn = func(_ string) (string, error) {
		hashCalled = true
		return "", nil
	}
	repoCalled := false
	f.userRepo.updateByUUIDFn = func(_ context.Context, _ string, _ domain.UpdateUser, _ *pgx.Tx) error {
		repoCalled = true
		return nil
	}

	err := f.useCase.UpdateByUUID(context.Background(), f.user.UUID, domain.UpdateUser{Login: "login", Password: "12345", Role: domain.UserRoleClient})
	if !errors.Is(err, domain.ErrShortPassword) {
		t.Fatalf("expected short password error, got=%v", err)
	}
	if hashCalled {
		t.Fatal("password hash must not be called for short password")
	}
	if repoCalled {
		t.Fatal("repo UpdateByUUID must not be called for short password")
	}
}

func Test_UserUseCase_DeleteByUUID_RepoError(t *testing.T) {
	f := newUserFixture()
	wantErr := errors.New("repo error")
	f.userRepo.deleteByUUIDFn = func(_ context.Context, _ string, _ *pgx.Tx) error {
		return wantErr
	}

	err := f.useCase.DeleteByUUID(context.Background(), f.user.UUID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got=%v", err)
	}
}

func Test_UserUseCase_DeleteByUUID_OK(t *testing.T) {
	f := newUserFixture()
	called := false
	f.userRepo.deleteByUUIDFn = func(_ context.Context, uuid string, tx *pgx.Tx) error {
		called = true
		if tx != nil {
			t.Fatal("expected nil tx")
		}
		if uuid != f.user.UUID {
			t.Fatalf("unexpected uuid: got=%q want=%q", uuid, f.user.UUID)
		}
		return nil
	}

	err := f.useCase.DeleteByUUID(context.Background(), f.user.UUID)
	if err != nil {
		t.Fatalf("DeleteByUUID returned error: %v", err)
	}
	if !called {
		t.Fatal("expected repo DeleteByUUID to be called")
	}
}
