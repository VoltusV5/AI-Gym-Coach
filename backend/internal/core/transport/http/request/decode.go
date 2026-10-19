// Package core_http_request декодирует JSON-запросы и проверяет входные данные.
package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "sport_app/internal/core/errors"

	"github.com/go-playground/validator/v10"
)

type validatable interface {
	Validate() error
}

var requestValidator = validator.New()

// DecodeAndValidateRequest декодирует тело r в dest и вызывает Validate, если метод реализован, иначе проверяет теги структуры через validator.
// При неверном JSON или нарушении правил возвращает ErrInvalidArgument.
func DecodeAndValidateRequest(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	var err error

	v, ok := dest.(validatable)
	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}

	if err != nil {
		return fmt.Errorf("request validation: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	return nil
}
