package app

import (
	"strings"
	"testing"

	"shc/config"
	"shc/internal/delivery/http"
	"shc/internal/infrastructure/cache"
	"shc/internal/repository"
	"shc/internal/service"
	"shc/internal/usecase"
)

func Test_App_checkMod_NilApp(t *testing.T) {
	var app *App

	err := app.checkMod()
	if err == nil || !strings.Contains(err.Error(), "app is nil") {
		t.Fatalf("expected nil app error, got %v", err)
	}
}

func Test_App_checkMod_MissingModules(t *testing.T) {
	app := &App{
		Repository: &repository.Repository{},
		cfg:        &config.Config{},
	}

	err := app.checkMod()
	if err == nil {
		t.Fatal("expected missing modules error")
	}
	for _, name := range []string{"Admin.Login", "Admin.Password", "MsgCache", "JWTService", "API"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("expected error to contain %q, got %v", name, err)
		}
	}
}

func Test_App_checkMod_OK(t *testing.T) {
	var encryptionKey [32]byte

	app := &App{
		cfg:               &config.Config{Admin: config.AdminConfig{Login: "admin", Password: "admin"}},
		Repository:        &repository.Repository{},
		MsgCache:          cache.NewMsgCache(),
		JWTService:        service.NewJWT("issuer", []byte("sign-key")),
		EncryptionService: service.NewAES256GCM(encryptionKey),
		AuthUseCase:       &usecase.AuthUseCase{},
		UserUseCase:       &usecase.UserUseCase{},
		ScheduleUseCase:   &usecase.ScheduleUseCase{},
		DepartmentUseCase: &usecase.DepartmentUseCase{},
		ChatUseCase:       &usecase.ChatUseCase{},
		API:               &http.API{},
	}

	if err := app.checkMod(); err != nil {
		t.Fatalf("checkMod returned error: %v", err)
	}
}
