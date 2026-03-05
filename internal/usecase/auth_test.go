package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"shc/domain"
	"shc/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeTx struct {
	commitErr     error
	rollbackErr   error
	commitCalls   int
	rollbackCalls int
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return nil, nil }
func (f *fakeTx) Commit(context.Context) error {
	f.commitCalls++
	return f.commitErr
}
func (f *fakeTx) Rollback(context.Context) error {
	f.rollbackCalls++
	return f.rollbackErr
}
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                         { return nil }

type mainRepoMock struct {
	createSessionFn func(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

func (m *mainRepoMock) CreateSession(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return m.createSessionFn(ctx, options)
}

type userRepoMock struct {
	getByUUIDFn  func(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error)
	getByLoginFn func(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error)
}

func (u *userRepoMock) GetByUUID(ctx context.Context, uuid string, tx *pgx.Tx) (*domain.User, error) {
	return u.getByUUIDFn(ctx, uuid, tx)
}

func (u *userRepoMock) GetByLogin(ctx context.Context, login string, tx *pgx.Tx) (*domain.User, error) {
	return u.getByLoginFn(ctx, login, tx)
}

type refreshTokenRepoMock struct {
	getFn              func(ctx context.Context, jti string, tx *pgx.Tx) (*domain.RefreshToken, error)
	createFn           func(ctx context.Context, jti, userUUID string, tx *pgx.Tx) error
	deleteByUserUUIDFn func(ctx context.Context, userUUID string, tx *pgx.Tx) error
	deleteByJTIFn      func(ctx context.Context, jti string, tx *pgx.Tx) error
}

func (r *refreshTokenRepoMock) Get(ctx context.Context, jti string, tx *pgx.Tx) (*domain.RefreshToken, error) {
	return r.getFn(ctx, jti, tx)
}

func (r *refreshTokenRepoMock) Create(ctx context.Context, jti, userUUID string, tx *pgx.Tx) error {
	return r.createFn(ctx, jti, userUUID, tx)
}

func (r *refreshTokenRepoMock) DeleteByUserUUID(ctx context.Context, userUUID string, tx *pgx.Tx) error {
	return r.deleteByUserUUIDFn(ctx, userUUID, tx)
}

func (r *refreshTokenRepoMock) DeleteByJTI(ctx context.Context, jti string, tx *pgx.Tx) error {
	return r.deleteByJTIFn(ctx, jti, tx)
}

type jwtServiceMock struct {
	createTokensFn       func(sub string, userRole domain.UserRole, scope []string, accessTokenTTL, refreshTokenTTL time.Duration) (tokens *domain.JWTTokens, refJTI string, err error)
	verifyAccessTokenFn  func(tokenStr string) (*service.AccessTokenClaims, error)
	verifyRefreshTokenFn func(tokenStr string) (*service.RefreshTokenClaims, error)
}

func (j *jwtServiceMock) CreateTokens(sub string, userRole domain.UserRole, scope []string, accessTokenTTL, refreshTokenTTL time.Duration) (tokens *domain.JWTTokens, refJTI string, err error) {
	return j.createTokensFn(sub, userRole, scope, accessTokenTTL, refreshTokenTTL)
}

func (j *jwtServiceMock) VerifyAccessToken(tokenStr string) (*service.AccessTokenClaims, error) {
	return j.verifyAccessTokenFn(tokenStr)
}

func (j *jwtServiceMock) VerifyRefreshToken(tokenStr string) (*service.RefreshTokenClaims, error) {
	return j.verifyRefreshTokenFn(tokenStr)
}

type pswdServiceMock struct {
	createPasswordHashFn func(password string) (string, error)
	verifyPasswordFn     func(password, hashedPassword string) bool
}

func (p *pswdServiceMock) CreatePasswordHash(password string) (string, error) {
	return p.createPasswordHashFn(password)
}

func (p *pswdServiceMock) VerifyPassword(password, hashedPassword string) bool {
	return p.verifyPasswordFn(password, hashedPassword)
}

type authFixture struct {
	useCase          *AuthUseCase
	tx               *fakeTx
	mainRepo         *mainRepoMock
	userRepo         *userRepoMock
	refreshTokenRepo *refreshTokenRepoMock
	jwtService       *jwtServiceMock
	pswdService      *pswdServiceMock
	user             *domain.User
	tokens           *domain.JWTTokens
}

func newAuthFixture() *authFixture {
	user := &domain.User{
		UUID:     "user-1",
		Login:    "login",
		Role:     domain.UserRoleManager,
		Password: "hashed-password",
	}
	tokens := &domain.JWTTokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}
	tx := &fakeTx{}

	mainRepo := &mainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return tx, nil
		},
	}
	userRepo := &userRepoMock{
		getByUUIDFn: func(_ context.Context, _ string, _ *pgx.Tx) (*domain.User, error) {
			return user, nil
		},
		getByLoginFn: func(_ context.Context, _ string, _ *pgx.Tx) (*domain.User, error) {
			return user, nil
		},
	}
	refreshTokenRepo := &refreshTokenRepoMock{
		getFn: func(_ context.Context, _ string, _ *pgx.Tx) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{JTI: "old-jti", UserUUID: user.UUID}, nil
		},
		createFn:           func(_ context.Context, _, _ string, _ *pgx.Tx) error { return nil },
		deleteByUserUUIDFn: func(_ context.Context, _ string, _ *pgx.Tx) error { return nil },
		deleteByJTIFn:      func(_ context.Context, _ string, _ *pgx.Tx) error { return nil },
	}
	jwtService := &jwtServiceMock{
		createTokensFn: func(_ string, _ domain.UserRole, _ []string, _, _ time.Duration) (*domain.JWTTokens, string, error) {
			return tokens, "new-jti", nil
		},
		verifyAccessTokenFn: func(_ string) (*service.AccessTokenClaims, error) {
			return &service.AccessTokenClaims{
				UserRole: user.Role,
				Scope:    []string{"tickets:read"},
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: user.UUID,
				},
			}, nil
		},
		verifyRefreshTokenFn: func(_ string) (*service.RefreshTokenClaims, error) {
			return &service.RefreshTokenClaims{
				JTI: "old-jti",
			}, nil
		},
	}
	pswdService := &pswdServiceMock{
		createPasswordHashFn: func(_ string) (string, error) { return "", nil },
		verifyPasswordFn:     func(_, _ string) bool { return true },
	}

	roleScopes := RoleScopes{
		domain.UserRoleManager: {"tickets:read"},
	}
	useCase := NewAuthUseCase(roleScopes, 2*time.Minute, 10*time.Minute, mainRepo, userRepo, refreshTokenRepo, jwtService, pswdService)

	return &authFixture{
		useCase:          useCase,
		tx:               tx,
		mainRepo:         mainRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		pswdService:      pswdService,
		user:             user,
		tokens:           tokens,
	}
}

func Test_NewAuthUseCase(t *testing.T) {
	f := newAuthFixture()
	if f.useCase == nil {
		t.Fatal("NewAuthUseCase returned nil")
	}
	if f.useCase.accessTokenTTL != 2*time.Minute {
		t.Fatalf("accessTokenTTL mismatch: got=%v", f.useCase.accessTokenTTL)
	}
	if f.useCase.refreshTokenTTL != 10*time.Minute {
		t.Fatalf("refreshTokenTTL mismatch: got=%v", f.useCase.refreshTokenTTL)
	}
}

func Test_AuthUseCase_Login_CreateSessionError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
}

func Test_AuthUseCase_Login_GetByLoginError(t *testing.T) {
	f := newAuthFixture()
	f.userRepo.getByLoginFn = func(_ context.Context, _ string, _ *pgx.Tx) (*domain.User, error) {
		return nil, errors.New("db error")
	}

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, domain.ErrLogin) {
		t.Fatalf("expected ErrLogin, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_PasswordMismatch(t *testing.T) {
	f := newAuthFixture()
	f.pswdService.verifyPasswordFn = func(_, _ string) bool { return false }

	tokens, err := f.useCase.Login(context.Background(), "login", "bad-password")
	if !errors.Is(err, domain.ErrLogin) {
		t.Fatalf("expected ErrLogin, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_CreateTokensError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create tokens error")
	f.jwtService.createTokensFn = func(_ string, _ domain.UserRole, _ []string, _, _ time.Duration) (*domain.JWTTokens, string, error) {
		return nil, "", wantErr
	}

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create tokens error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_DeleteByUserUUIDError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("delete refresh error")
	f.refreshTokenRepo.deleteByUserUUIDFn = func(_ context.Context, _ string, _ *pgx.Tx) error {
		return wantErr
	}

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete refresh error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_CreateRefreshTokenError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create refresh error")
	f.refreshTokenRepo.createFn = func(_ context.Context, _, _ string, _ *pgx.Tx) error {
		return wantErr
	}

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create refresh error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_CommitError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit to be called once, got=%d", f.tx.commitCalls)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Login_OK(t *testing.T) {
	f := newAuthFixture()

	tokens, err := f.useCase.Login(context.Background(), "login", "password")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if tokens == nil {
		t.Fatal("Login returned nil tokens")
	}
	if tokens.AccessToken != f.tokens.AccessToken || tokens.RefreshToken != f.tokens.RefreshToken {
		t.Fatalf("unexpected tokens: %+v", tokens)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit to be called once, got=%d", f.tx.commitCalls)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Logout_VerifyAccessTokenError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("verify access error")
	f.jwtService.verifyAccessTokenFn = func(_ string) (*service.AccessTokenClaims, error) {
		return nil, wantErr
	}

	err := f.useCase.Logout(context.Background(), "access-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected verify access error, got=%v", err)
	}
}

func Test_AuthUseCase_Logout_DeleteByUserUUIDError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("delete by user uuid error")
	f.refreshTokenRepo.deleteByUserUUIDFn = func(_ context.Context, _ string, tx *pgx.Tx) error {
		if tx != nil {
			t.Fatalf("expected nil tx in Logout delete, got=%v", tx)
		}
		return wantErr
	}

	err := f.useCase.Logout(context.Background(), "access-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete by user uuid error, got=%v", err)
	}
}

func Test_AuthUseCase_Logout_OK(t *testing.T) {
	f := newAuthFixture()

	err := f.useCase.Logout(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
}

func Test_AuthUseCase_GetAccessTokenClaims(t *testing.T) {
	f := newAuthFixture()
	want := &service.AccessTokenClaims{
		UserRole: domain.UserRoleManager,
		Scope:    []string{"tickets:read"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-1",
		},
	}
	f.jwtService.verifyAccessTokenFn = func(_ string) (*service.AccessTokenClaims, error) {
		return want, nil
	}

	claims, err := f.useCase.GetAccessTokenClaims("access-token")
	if err != nil {
		t.Fatalf("GetAccessTokenClaims returned error: %v", err)
	}
	if claims != want {
		t.Fatalf("unexpected claims pointer: got=%p want=%p", claims, want)
	}
}

func Test_AuthUseCase_Refresh_VerifyRefreshTokenError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("verify refresh error")
	f.jwtService.verifyRefreshTokenFn = func(_ string) (*service.RefreshTokenClaims, error) {
		return nil, wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected verify refresh error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
}

func Test_AuthUseCase_Refresh_CreateSessionError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
}

func Test_AuthUseCase_Refresh_GetFromDBError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("get refresh from db error")
	f.refreshTokenRepo.getFn = func(_ context.Context, _ string, _ *pgx.Tx) (*domain.RefreshToken, error) {
		return nil, wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected get refresh from db error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_GetByUUIDError(t *testing.T) {
	f := newAuthFixture()
	f.userRepo.getByUUIDFn = func(_ context.Context, _ string, _ *pgx.Tx) (*domain.User, error) {
		return nil, errors.New("get user by uuid error")
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, domain.ErrLogin) {
		t.Fatalf("expected ErrLogin, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_CreateTokensError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create tokens error")
	f.jwtService.createTokensFn = func(_ string, _ domain.UserRole, _ []string, _, _ time.Duration) (*domain.JWTTokens, string, error) {
		return nil, "", wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create tokens error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_DeleteByUserUUIDError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("delete by user uuid error")
	f.refreshTokenRepo.deleteByUserUUIDFn = func(_ context.Context, _ string, _ *pgx.Tx) error {
		return wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete by user uuid error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_CreateRefreshTokenError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("create refresh token error")
	f.refreshTokenRepo.createFn = func(_ context.Context, _, _ string, _ *pgx.Tx) error {
		return wantErr
	}

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create refresh token error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_CommitError(t *testing.T) {
	f := newAuthFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
	if tokens != nil {
		t.Fatalf("expected nil tokens on error, got=%+v", tokens)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit to be called once, got=%d", f.tx.commitCalls)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}

func Test_AuthUseCase_Refresh_OK(t *testing.T) {
	f := newAuthFixture()

	tokens, err := f.useCase.Refresh(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if tokens == nil {
		t.Fatal("Refresh returned nil tokens")
	}
	if tokens.AccessToken != f.tokens.AccessToken || tokens.RefreshToken != f.tokens.RefreshToken {
		t.Fatalf("unexpected tokens: %+v", tokens)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit to be called once, got=%d", f.tx.commitCalls)
	}
	if f.tx.rollbackCalls != 1 {
		t.Fatalf("expected rollback to be called once, got=%d", f.tx.rollbackCalls)
	}
}
