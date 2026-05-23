package domain

import "errors"

var (
	ErrInvalidTimeFormat    error = errors.New("Параметры from и to имеют неверный формат (пример: 08:00, 18:00)")
	ErrInvalidDayType       error = errors.New("Значение True может стоять только у is_break или тольок у is_week_end")
	ErrLogin                error = errors.New("Логи или пароль некорректны")
	ErrShortPassword        error = errors.New("Пароль должен быть равен или длинее 6 символов")
	ErrLimitIsBiggerThen100 error = errors.New("Лимит должен быть меньше или равен 100")
	ErrDurationFromTo       error = errors.New("Максимальная разница между датой начала(from) и датой конца(to) не должна быть больше чем 31 день")
	ErrFilesLoad            error = errors.New("Максимальное количество файлов для загрузки 10, максимальный обем одного файла 15МБ")
	ErrUnknownObject        error = errors.New("Неизвестный объект")
)
