package app

import (
	"context"
	"log/slog"

	"shc/config"
	"shc/domain"
	"shc/internal/delivery/http"
	"shc/internal/infrastructure/cache"
	"shc/internal/infrastructure/oneass"
	"shc/internal/infrastructure/s3"
	"shc/internal/repository"
	"shc/internal/service"
	"shc/internal/usecase"

	"github.com/jackc/pgx/v5"
)

type App struct {
	cfg               *config.Config
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
	initLogger()

	passwordService := service.PasswordCoder{}
	msgCache := cache.NewMsgCache()
	jwtService := service.NewJWT(cfg.JWT.Issuer, []byte(cfg.JWT.SignKey))
	encryptionService := service.NewAES256GCM(cfg.Encryption.Key32)

	baseRepo, err := repository.NewRepository(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	s3Storage, err := s3.NewMiniO(ctx, cfg.MiniO.URL, cfg.MiniO.Login, cfg.MiniO.Password, cfg.MiniO.PrivateBucket)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepo(baseRepo)
	refreshTokenRepo := repository.NewRefreshTokenRepo(baseRepo)
	departmentRepo := repository.NewDepartmentRepo(baseRepo)
	scheduleRepo := repository.NewScheduleRepo(baseRepo)
	msgRepo := repository.NewMsgRepo(baseRepo)
	// ticketRepo := repository.NewTicketRepo(baseRepo)
	otherSystemLogin := oneass.NewOneAssClientLogin()

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
	userUseCase := usecase.NewUserUseCase(baseRepo, userRepo, passwordService)
	// ticketUseCase := usecase.NewTicketUseCase(ticketRepo)
	departmentUseCase := usecase.NewDepartmentUseCase(baseRepo, departmentRepo, scheduleUseCase, scheduleRepo)
	chatUseCase := usecase.NewChatUseCase(msgCache, s3Storage, msgRepo, baseRepo, encryptionService, cfg.Chat.ReadTimeout, cfg.Chat.PollInterval)
	api := http.NewAPI(authUseCase, cfg.HTTP)

	app := App{
		cfg:               &cfg,
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
	}

	if err := app.checkMod(); err != nil {
		return nil, err
	}

	return &app, nil
}

func (a *App) Run(ctx context.Context) error {
	admin, err := a.UserUseCase.GetByLogin(ctx, a.cfg.Admin.Login)
	if err == pgx.ErrNoRows {
		slog.Info("Админ создан")
		if err := a.UserUseCase.CreateAdmin(ctx,
			domain.UserSystemInfo{UserRole: domain.UserRoleAdmin},
			domain.CreateUser{
				Login:    a.cfg.Admin.Login,
				Password: a.cfg.Admin.Password,
				Role:     domain.UserRoleAdmin,
			}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		slog.Info("Пароль админа изменен")
		if err := a.UserUseCase.UpdateByUUID(ctx,
			domain.UserSystemInfo{UserRole: domain.UserRoleAdmin},
			admin.UUID,
			domain.UpdateUser{
				Password: a.cfg.Admin.Password,
			}); err != nil {
			return err
		}
	}

	slog.Info("APP is START")
	return a.API.Run()
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.API.Shutdown(ctx); err != nil {
		return err
	}

	a.Repository.Close()
	slog.Info("APP is SHUTDOWN")
	return nil
}
