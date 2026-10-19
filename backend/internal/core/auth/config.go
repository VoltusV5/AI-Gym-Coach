// Package core_auth реализует создание и проверку JWT и загрузку настроек авторизации.
package core_auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config содержит настройки JWT из окружения: обязательный JWT_SECRET для подписи
// и JWT_TOKEN_TTL со сроком действия токена (по умолчанию 2160h).
type Config struct {
	Secret   string        `envconfig:"SECRET" required:"true"`
	TokenTTL time.Duration `envconfig:"TOKEN_TTL" default:"2160h"`
}

// NewConfig загружает настройки с префиксом JWT из окружения.
// Возвращает ошибку при отсутствии обязательного значения или неверном формате.
func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("JWT", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

// NewConfigMust загружает настройки из окружения и вызывает панику при ошибке конфигурации.
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get JWT config: %w", err)
		panic(err)
	}

	return config
}
