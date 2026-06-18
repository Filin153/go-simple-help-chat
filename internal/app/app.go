package app

import (
	"context"

	"shc/config"
	"shc/domain"
	"shc/internal/delivery/http"
	"shc/internal/infrastructure/cache"
	"shc/internal/repository"
	"shc/internal/service"
	"shc/internal/usecase"
)

var newRepository = repository.NewRepository

type App struct {
	Repository        *repository.Repository
	MsgCache          *cache.MsgCache
	JWTService        *service.JWT
	EncryptionService *service.AES256GCM
	AuthUseCase       *usecase.AuthUseCase
	UserUseCase       *usecase.UserUseCase
	ScheduleUseCase   *usecase.ScheduleUseCase
	DepartmentUseCase *usecase.DepartmentUseCase
	ChatUseCase       *usecase.ChatUseCase
	API               *http.API
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	cfg = config.ApplyDefaults(cfg)

	passwordService := service.PasswordCoder{}
	msgCache := cache.NewMsgCache()
	jwtService := service.NewJWT(cfg.JWT.Issuer, []byte(cfg.JWT.SignKey))
	encryptionService := service.NewAES256GCM(cfg.Encryption.Key32)

	baseRepo, err := newRepository(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepo(baseRepo)
	refreshTokenRepo := repository.NewRefreshTokenRepo(baseRepo)
	departmentRepo := repository.NewDepartmentRepo(baseRepo)
	scheduleRepo := repository.NewScheduleRepo(baseRepo)
	msgRepo := repository.NewMsgRepo(baseRepo)
	ticketRepo := repository.NewTicketRepo(baseRepo)
	s3 := stubS3{}
	otherSystemLogin := stubOtherSystemLogin{}

	scheduleUseCase := usecase.NewScheduleUseCase(baseRepo, scheduleRepo)
	authUseCase := usecase.NewAuthUseCase(
		usecase.RoleScopes(cfg.Auth.RoleScopes),
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		baseRepo,
		userRepo,
		refreshTokenRepo,
		jwtService,
		passwordService,
		otherSystemLogin,
	)
	userUseCase := usecase.NewUserUseCase(userRepo, passwordService)
	departmentUseCase := usecase.NewDepartmentUseCase(baseRepo, departmentRepo, scheduleUseCase, scheduleRepo)
	chatUseCase := usecase.NewChatUseCase(msgCache, s3, msgRepo, ticketRepo, baseRepo, encryptionService, cfg.Chat.ReadTimeout, cfg.Chat.PollInterval)
	api := http.NewAPI(authHTTPAdapter{auth: authUseCase}, cfg.HTTP)

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

type authHTTPAdapter struct {
	auth authUseCase
}

type authUseCase interface {
	Login(ctx context.Context, login, password string) (*domain.JWTTokens, error)
	LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error)
	Logout(ctx context.Context, accessToken string) error
	GetAccessTokenClaims(accessToken string) (*service.AccessTokenClaims, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error)
}

func (a authHTTPAdapter) Login(ctx context.Context, login, password string) (*domain.JWTTokens, error) {
	return a.auth.Login(ctx, login, password)
}

func (a authHTTPAdapter) LoginClient(ctx context.Context, args ...any) (*domain.JWTTokens, error) {
	return a.auth.LoginClient(ctx, args...)
}

func (a authHTTPAdapter) Logout(ctx context.Context, accessToken string) error {
	return a.auth.Logout(ctx, accessToken)
}

func (a authHTTPAdapter) GetAccessTokenClaims(ctx context.Context, accessToken string) error {
	_, err := a.auth.GetAccessTokenClaims(accessToken)
	return err
}

func (a authHTTPAdapter) Refresh(ctx context.Context, refreshToken string) (*domain.JWTTokens, error) {
	return a.auth.Refresh(ctx, refreshToken)
}

type stubS3 struct{}

func (stubS3) Save(context.Context, string, []byte) (string, error) {
	return "", domain.ErrUnknownObject
}

func (stubS3) Delete(context.Context, string) error {
	return domain.ErrUnknownObject
}

type stubOtherSystemLogin struct{}

func (stubOtherSystemLogin) Login(context.Context, ...any) (string, map[any]any, error) {
	return "", nil, domain.ErrUnknownObject
}
