package domain

import "errors"

var (
	ErrInvalidTimeFormat error = errors.New("Параметры from и to имеют неверный формат (пример: 08:00, 18:00)")
	ErrLogin             error = errors.New("Логи или пароль некорректны")
	ErrShortPassword     error = errors.New("Пароль должен быть равен или длинее 6 символов")
)
