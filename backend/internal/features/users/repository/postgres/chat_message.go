package users_postgres_repository

import (
	"context"
	"fmt"
	"time"
)

func (r *UsersRepository) InsertChatMessage(
	ctx context.Context,
	userID string,
	role string,
	content string,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sportapp.chat_messages (user_id, role, content, send_at)
		VALUES ($1, $2, $3, $4)
	`, userID, role, content, time.Now())
	if err != nil {
		return fmt.Errorf("insert chat_messages: %w", err)
	}

	return nil
}
