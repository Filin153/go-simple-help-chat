package cmd

import (
	"context"
	"errors"
	"time"

	"shc/config"
	"shc/domain"
	deliveryhttp "shc/internal/delivery/http"
	infraCache "shc/internal/infrastructure/cache"
	"shc/internal/repository"
	appservice "shc/internal/service"
	"shc/internal/usecase"

	"github.com/jackc/pgx/v5"
)

var (
	errInfraNotConfigured = errors.New("infrastructure is not configured")
)

type App struct {
	Repository        *repository.Repository
	MsgCache          *infraCache.MsgCache
	JWTService        *appservice.JWT
	EncryptionService *appservice.AES256GCM
	AuthUseCase       *usecase.AuthUseCase
	UserUseCase       *usecase.UserUseCase
	ScheduleUseCase   *usecase.ScheduleUseCase
	DepartmentUseCase *usecase.DepartmentUseCase
	ChatUseCase       *usecase.ChatUseCase
	API               *deliveryhttp.API
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	cfg = config.ApplyDefaults(cfg)

	passwordService := passwordServiceAdapter{}
	msgCache := infraCache.NewMsgCache()
	jwtService := appservice.NewJWT(cfg.JWT.Issuer, []byte(cfg.JWT.SignKey))
	encryptionService := appservice.NewAES256GCM(cfg.Encryption.Key32)

	var baseRepo *repository.Repository
	var mainRepo usecase.MainRepo = stubMainRepo{}
	if cfg.PostgresDSN != "" {
		repo, err := repository.NewRepository(ctx, cfg.PostgresDSN)
		if err != nil {
			return nil, err
		}
		baseRepo = repo
		mainRepo = repo
	}

	userRepo := stubUserRepo{}
	refreshTokenRepo := stubRefreshTokenRepo{}
	departmentRepo := stubDepartmentRepo{}
	scheduleRepo := stubScheduleRepo{}
	msgRepo := stubMsgRepo{}
	ticketRepo := stubTicketRepo{}
	s3 := stubS3{}
	otherSystemLogin := stubOtherSystemLogin{}

	scheduleUseCase := usecase.NewScheduleUseCase(mainRepo, scheduleRepo)
	authUseCase := usecase.NewAuthUseCase(
		usecase.RoleScopes(cfg.Auth.RoleScopes),
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		mainRepo,
		userRepo,
		refreshTokenRepo,
		jwtService,
		passwordService,
		otherSystemLogin,
	)
	userUseCase := usecase.NewUserUseCase(userRepo, passwordService)
	departmentUseCase := usecase.NewDepartmentUseCase(mainRepo, departmentRepo, scheduleUseCase, scheduleRepo)
	chatUseCase := usecase.NewChatUseCase(msgCache, s3, msgRepo, ticketRepo, mainRepo, encryptionService, cfg.Chat.ReadTimeout, cfg.Chat.PollInterval)
	api := deliveryhttp.NewAPI(httpAuthAdapter{auth: authUseCase}, cfg.HTTP)

	return &App{
		Repository:        baseRepo,
		MsgCache:          msgCache,
		JWTService:        jwtService,
		EncryptionService: encryptionService,
		AuthUseCase:       authUseCase,
		UserUseCase:       userUseCase,
		ScheduleUseCase:   scheduleUseCase,
		DepartmentUseCase: departmentUseCase,
		ChatUseCase:       chatUseCase,
		API:               api,
	}, nil
}

type httpAuthAdapter struct {
	auth *usecase.AuthUseCase
}

func (a httpAuthAdapter) Login(ctx context.Context, login, password string) (*domain.JWTTokens, error) {
	return a.auth.Login(ctx, login, password)
}

func (a httpAuthAdapter) LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error) {
	return a.auth.LoginClient(ctx, args...)
}

func (a httpAuthAdapter) Logout(ctx context.Context, accessToken string) error {
	return a.auth.Logout(ctx, accessToken)
}

func (a httpAuthAdapter) GetAccessTokenClaims(_ context.Context, accessToken string) error {
	_, err := a.auth.GetAccessTokenClaims(accessToken)
	return err
}

func (a httpAuthAdapter) Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error) {
	return a.auth.Refresh(ctx, refreshToken)
}

type passwordServiceAdapter struct{}

func (passwordServiceAdapter) CreatePasswordHash(password string) (string, error) {
	return appservice.CreatePasswordHash(password)
}

func (passwordServiceAdapter) VerifyPassword(password, hashedPassword string) bool {
	return appservice.VerifyPassword(password, hashedPassword)
}

type stubMainRepo struct{}

func (stubMainRepo) CreateSession(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errInfraNotConfigured
}

type stubUserRepo struct{}

func (stubUserRepo) GetAll(context.Context, pgx.Tx) ([]domain.User, error) {
	return nil, errInfraNotConfigured
}

func (stubUserRepo) GetByUUID(context.Context, string, pgx.Tx) (*domain.User, error) {
	return nil, errInfraNotConfigured
}

func (stubUserRepo) GetByLogin(context.Context, string, pgx.Tx) (*domain.User, error) {
	return nil, errInfraNotConfigured
}

func (stubUserRepo) Create(context.Context, domain.CreateUser, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubUserRepo) UpdateByUUID(context.Context, string, domain.UpdateUser, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubUserRepo) DeleteByUUID(context.Context, string, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubUserRepo) UpdateUserPasswordByUUID(context.Context, string, string, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubUserRepo) CreateClient(context.Context, domain.CreateClient, pgx.Tx) error {
	return errInfraNotConfigured
}

type stubRefreshTokenRepo struct{}

func (stubRefreshTokenRepo) Get(context.Context, string, pgx.Tx) (*domain.RefreshToken, error) {
	return nil, errInfraNotConfigured
}

func (stubRefreshTokenRepo) Create(context.Context, string, string, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubRefreshTokenRepo) DeleteByUserUUID(context.Context, string, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubRefreshTokenRepo) DeleteByJTI(context.Context, string, pgx.Tx) error {
	return errInfraNotConfigured
}

type stubDepartmentRepo struct{}

func (stubDepartmentRepo) Create(context.Context, domain.CreateDepartment, pgx.Tx) (int, error) {
	return 0, errInfraNotConfigured
}

func (stubDepartmentRepo) GetAll(context.Context, int, int, int, pgx.Tx) ([]domain.DepartmentWithOnScheduleeDay, error) {
	return nil, errInfraNotConfigured
}

func (stubDepartmentRepo) GetByID(context.Context, int, pgx.Tx) (domain.Department, error) {
	return domain.Department{}, errInfraNotConfigured
}

func (stubDepartmentRepo) Update(context.Context, int, domain.UpdateDepartment, pgx.Tx) error {
	return errInfraNotConfigured
}

type stubScheduleRepo struct{}

func (stubScheduleRepo) Create(context.Context, *domain.CreateSchedule, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubScheduleRepo) Update(context.Context, *domain.UpdateSchedule, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubScheduleRepo) Delete(context.Context, int, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubScheduleRepo) ExistByDepartmentID(context.Context, int) (bool, error) {
	return false, errInfraNotConfigured
}

func (stubScheduleRepo) GetFromTo(context.Context, int, time.Time, time.Time, pgx.Tx) ([]domain.Schedule, error) {
	return nil, errInfraNotConfigured
}

type stubMsgRepo struct{}

func (stubMsgRepo) CreateForClient(context.Context, domain.CreateMsg, pgx.Tx) (*domain.Msg, error) {
	return nil, errInfraNotConfigured
}

func (stubMsgRepo) CreateForManager(context.Context, domain.CreateMsg, pgx.Tx) (*domain.Msg, error) {
	return nil, errInfraNotConfigured
}

func (stubMsgRepo) CreateFile(context.Context, int, string, string, pgx.Tx) (domain.MsgFileContent, error) {
	return domain.MsgFileContent{}, errInfraNotConfigured
}

func (stubMsgRepo) GetUnread(context.Context, string, pgx.Tx) ([]*domain.Msg, error) {
	return nil, errInfraNotConfigured
}

func (stubMsgRepo) MarkReadByID(context.Context, string, int, pgx.Tx) error {
	return errInfraNotConfigured
}

func (stubMsgRepo) GetHistory(context.Context, string, int, time.Time, time.Time) ([]*domain.Msg, error) {
	return nil, errInfraNotConfigured
}

type stubTicketRepo struct{}

func (stubTicketRepo) GetByID(context.Context, int) (domain.Ticket, error) {
	return domain.Ticket{}, errInfraNotConfigured
}

type stubS3 struct{}

func (stubS3) Save(context.Context, string, []byte) (string, error) {
	return "", errInfraNotConfigured
}

func (stubS3) Delete(context.Context, string) error {
	return errInfraNotConfigured
}

type stubOtherSystemLogin struct{}

func (stubOtherSystemLogin) Login(context.Context, ...any) (string, map[any]any, error) {
	return "", nil, errInfraNotConfigured
}
