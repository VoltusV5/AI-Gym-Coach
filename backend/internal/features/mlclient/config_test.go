package mlclient

import (
	"testing"
	"time"
)

func TestTimeoutConfiguration(t *testing.T) {
	t.Setenv("ML_BASE_URL", "http://ml:8000")
	t.Setenv("ML_TIMEOUT", "7s")
	t.Setenv("ML_CHAT_TIMEOUT", "19s")
	t.Setenv("ML_GENERATION_TIMEOUT", "23s")
	cfg, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 7*time.Second || cfg.ChatRequestTimeout() != 19*time.Second || cfg.PlanGenerationTimeout() != 23*time.Second {
		t.Fatalf("таймауты окружения не применились: %+v", cfg)
	}
	if (Config{Timeout: 7 * time.Second}).ChatRequestTimeout() != 7*time.Second {
		t.Fatal("явный общий таймаут должен работать для программно созданной конфигурации")
	}
	if (Config{}).ChatRequestTimeout() != 60*time.Second || (Config{}).PlanGenerationTimeout() != 52*time.Second {
		t.Fatal("нарушены значения по умолчанию")
	}
}
