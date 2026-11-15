package core_http_middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	"strings"
)

func AccountStatus(jwt *core_auth.JWT, pool core_postgres_pool.Pool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.Fields(r.Header.Get("Authorization"))
			if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := jwt.ParseToken(header[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), pool.OpTimeout())
			defer cancel()
			var blocked bool
			err = pool.QueryRow(ctx, `SELECT blocked FROM sportapp.users WHERE id=$1`, claims.UserID).Scan(&blocked)
			if err == nil && !blocked {
				next.ServeHTTP(w, r)
				return
			}
			if errors.Is(err, core_postgres_pool.ErrNoRows) {
				err = fmt.Errorf("account missing: %w", core_errors.ErrUnauthorized)
			}
			if err == nil {
				err = fmt.Errorf("account blocked: %w", core_errors.ErrForbidden)
			}
			response := core_http_responce.NewHTTPResponce(core_logger.FromContext(r.Context()), w)
			if strings.HasPrefix(r.URL.Path, "/api/v2/") {
				response = core_http_responce.NewHTTPResponceV2(core_logger.FromContext(r.Context()), w, r.Header.Get("X-Request-ID"))
			}
			response.ErrorResponse(err, "account status")
		})
	}
}
