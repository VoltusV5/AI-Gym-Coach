// Package core_auth provides JWT authentication utilities, token parsing, and configuration loading.
package core_auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds JWT authentication configuration parameters loaded from environment variables.
//
// Fields:
//   - Secret: Shared secret key used for signing and verifying HMAC-SHA256 tokens (env: JWT_SECRET, required).
//   - TokenTTL: Duration for which minted JWT tokens remain valid before expiring (env: JWT_TOKEN_TTL, default: 2160h).
type Config struct {
	Secret   string        `envconfig:"SECRET" required:"true"`
	TokenTTL time.Duration `envconfig:"TOKEN_TTL" default:"2160h"`
}

// NewConfig parses and validates JWT configuration from environment variables prefixed with "JWT".
//
// Returns:
//   - Config: Populated configuration struct.
//   - error: Non-nil if required fields are missing or format is invalid.
func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("JWT", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

// NewConfigMust parses JWT configuration from environment variables and panics if parsing fails.
//
// Returns:
//   - Config: Validated configuration struct.
//
// Side effects:
//   - Panics if configuration is invalid or missing required variables.
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get JWT config: %w", err)
		panic(err)
	}

	return config
}
