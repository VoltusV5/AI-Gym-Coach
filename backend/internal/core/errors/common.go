// Package core_errors определяет общие ошибки приложения для согласованной обработки в HTTP, сервисах и тестах.
package core_errors

import "errors"

var (
	// ErrNotFound означает, что объект не существует, удалён или недоступен пользователю; соответствует HTTP 404.
	ErrNotFound = errors.New("not found")

	// ErrInvalidArgument означает неверные параметры или тело запроса; соответствует HTTP 400.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrConflict означает конфликт версии или уникальности; соответствует HTTP 409.
	ErrConflict = errors.New("conflict")

	// ErrUnauthorized означает отсутствие или недействительность учётных данных; соответствует HTTP 401.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTooManyRequests означает превышение квоты запросов; соответствует HTTP 429.
	ErrTooManyRequests = errors.New("too many requests")
)
