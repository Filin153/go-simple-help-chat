package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"shc/config"
	"shc/internal/app"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	downCTX, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Run(ctx); err != nil {
			slog.Error("Ошибка сервера", "err", err)
			log.Fatalf("Ошибка сервера: %v\n", err)
		}
	}()

	<-downCTX.Done()

	log.Println("Получен сигнал завершения. Начинаем Graceful Shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
		slog.Error("Ошибка при остановке сервера", "err", err)
		log.Fatalf("Ошибка при остановке сервера: %v\n", err)
	}

	slog.Info("Сервер успешно остановлен")
}
