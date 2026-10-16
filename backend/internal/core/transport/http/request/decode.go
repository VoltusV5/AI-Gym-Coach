// Package core_http_request provides JSON request deserialization and schema validation.
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

// DecodeAndValidateRequest unmarshals JSON request body into the dest pointer, and executes
// either custom Validate() method (if implemented) or struct tag validation via go-playground/validator.
//
// Parameters:
//   - r: Incoming HTTP request containing JSON body.
//   - dest: Pointer to target struct.
//
// Returns:
//   - error: core_errors.ErrInvalidArgument if JSON decoding or validation fails.
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
