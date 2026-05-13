package aichat

import (
	"fmt"
	"strings"

	core_errors "sport_app/internal/core/errors"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateAnswerRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

func (r *GenerateAnswerRequest) Validate() error {
	if strings.TrimSpace(r.Model) == "" {
		return fmt.Errorf("`model` is required: %w", core_errors.ErrInvalidArgument)
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("`messages` must not be empty: %w", core_errors.ErrInvalidArgument)
	}
	for i, m := range r.Messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role != "user" && role != "assistant" {
			return fmt.Errorf(
				"`messages[%d].role` must be user or assistant: %w",
				i,
				core_errors.ErrInvalidArgument,
			)
		}
		if strings.TrimSpace(m.Content) == "" {
			return fmt.Errorf(
				"`messages[%d].content` must not be empty: %w",
				i,
				core_errors.ErrInvalidArgument,
			)
		}
	}
	return nil
}

func normalizeMessages(in []ChatMessage) []ChatMessage {
	out := make([]ChatMessage, len(in))
	for i, m := range in {
		out[i] = ChatMessage{
			Role:    strings.ToLower(strings.TrimSpace(m.Role)),
			Content: m.Content,
		}
	}
	return out
}

func lastUserContent(msgs []ChatMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(msgs[i].Role), "user") {
			return msgs[i].Content
		}
	}
	return ""
}