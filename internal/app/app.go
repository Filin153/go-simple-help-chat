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

	baseRepo, err := repository.NewRepository(ctx, cfg.PostgresDSN)
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

