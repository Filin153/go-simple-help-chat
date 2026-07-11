package app

import (
	"log/slog"
	"os"
)

func initLogger() {
	// 1. Настраиваем опции (например, уровень логирования)
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo, // Включаем Debug логи
		AddSource: true,
	}

	// 2. Создаем нужный обработчик: JSON для продакшена или Text для локальной разработки
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// 3. Создаем экземпляр логгера
	logger := slog.New(handler)

	// 4. Делаем его глобальным по умолчанию
	slog.SetDefault(logger)
}
