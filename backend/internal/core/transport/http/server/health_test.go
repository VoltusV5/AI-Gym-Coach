package core_http_server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_logger "sport_app/internal/core/logger"
)

func TestHealthCheck(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "ready", true: "database_unavailable"}[failed], func(t *testing.T) {
			s := NewHTTPServer(Config{}, core_logger.Nop())
			s.RegisterHealthCheck(func(ctx context.Context) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("проверка БД должна иметь таймаут")
				}
				if failed {
					return errors.New("database password must not appear in response")
				}
				return nil
			})
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			want := http.StatusOK
			if failed {
				want = http.StatusServiceUnavailable
			}
			if w.Code != want || strings.Contains(w.Body.String(), "password") {
				t.Fatalf("неверный ответ готовности: %d %q", w.Code, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("результат проверки готовности не должен кешироваться")
			}
		})
	}
}
