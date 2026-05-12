package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"go.uber.org/zap"
)

const (
	maxAssistantStoredBytes = 2 << 20
	chatUserRatePerMinute   = 30
)

type UsersReader interface {
	GetProfile(ctx context.Context, userID string) (users_postgres_repository.Profile, error)
	GetPlanTemplateJSON(ctx context.Context, userID string) (json.RawMessage, error)
	InsertChatMessage(ctx context.Context, userID string, role string, content string) error
	GetChatMessages(ctx context.Context, userID string, limit int) ([]users_postgres_repository.ChatMessage, error)
}

type Service struct {
	cfg        mlclient.Config
	httpClient *http.Client
	users      UsersReader
}

func NewService(cfg mlclient.Config, users UsersReader) *Service {
	return &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   0,
		},
		users: users,
	}
}

func (s *Service) streamPython(
	ctx context.Context,
	body []byte,
) (*http.Response, error) {
	u := s.cfg.ChatURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build python chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream, application/json, text/plain, */*")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("python chat request: %w", err)
	}
	return resp, nil
}

func (s *Service) generateAnswer(
	ctx context.Context,
	userID string,
	req GenerateAnswerRequest,
	rw http.ResponseWriter,
) error {
	planJSON, err := s.users.GetPlanTemplateJSON(ctx, userID)
	if err != nil {
		return fmt.Errorf("plan template: %w", err)
	}

	profile, err := s.users.GetProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile: %w", err)
	}

	pyBody, err := buildPythonChatBody(req.Model, req.Messages, planJSON, profile)
	if err != nil {
		return fmt.Errorf("build python body: %w", err)
	}

	if u := lastUserContent(req.Messages); u != "" {
		if err := s.users.InsertChatMessage(ctx, userID, "user", u); err != nil {
			return fmt.Errorf("store user message: %w", err)
		}
	}

	resp, err := s.streamPython(ctx, pyBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("python chat status %d: %s", resp.StatusCode, string(b))
	}

	fl, ok := rw.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		rw.Header().Set("Content-Type", ct)
	} else {
		rw.Header().Set("Content-Type", "text/event-stream")
	}
	if v := resp.Header.Get("Cache-Control"); v != "" {
		rw.Header().Set("Cache-Control", v)
	} else {
		rw.Header().Set("Cache-Control", "no-cache")
	}
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no")
	rw.WriteHeader(http.StatusOK)

	log := core_logger.FromContext(ctx)
	var assistantBuf bytes.Buffer
	chunk := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(chunk)
		if n > 0 {
			if _, werr := rw.Write(chunk[:n]); werr != nil {
				log.Warn("ai_chat client write", zap.Error(werr))
				assistantBuf.Write(chunk[:n])
				break
			}
			fl.Flush()
			if assistantBuf.Len() < maxAssistantStoredBytes {
				space := maxAssistantStoredBytes - assistantBuf.Len()
				if space > n {
					space = n
				}
				if space > 0 {
					assistantBuf.Write(chunk[:space])
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			log.Warn("ai_chat upstream read", zap.Error(readErr))
			break
		}
	}

	if assistantBuf.Len() > 0 {
		if err := s.users.InsertChatMessage(ctx, userID, "assistant", assistantBuf.String()); err != nil {
			log.Warn("ai_chat store assistant", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) HandleGenerateAnswer(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var req GenerateAnswerRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode ai_chat body")
		return
	}

	log.Info(
		"ai_chat generate_answer",
		zap.String("user_id", userID),
		zap.String("model", req.Model),
		zap.Int("messages_count", len(req.Messages)),
	)

	if err := s.generateAnswer(ctx, userID, req, rw); err != nil {
		responseHandler.ErrorResponse(err, "ai chat failed")
	}
}

func (s *Service) HandleGetHistory(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	msgs, err := s.users.GetChatMessages(ctx, userID, 50)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get chat history")
		return
	}

	// Convert repository ChatMessage to local ChatMessage struct for JSON rendering
	res := make([]ChatMessage, len(msgs))
	for i, m := range msgs {
		res[i] = ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	responseHandler.JSONResponse(res, http.StatusOK)
}

func (s *Service) RegisterRoutes(
	jwt *core_auth.JWT,
	register func(method, path string, handler http.Handler),
) {
	protect := core_http_middleware.Protect(jwt)
	rl := core_http_middleware.UserRateLimit(chatUserRatePerMinute, time.Minute)

	hGen := protect(rl(http.HandlerFunc(s.HandleGenerateAnswer)))
	register(http.MethodPost, "/ai_chat/generate_answer", hGen)

	hHist := protect(http.HandlerFunc(s.HandleGetHistory))
	register(http.MethodGet, "/ai_chat/history", hHist)
}
