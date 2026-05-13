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

func (r *UsersRepository) GetChatMessages(
	ctx context.Context,
	userID string,
	limit int,
) ([]ChatMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT role, content
		FROM (
			SELECT role, content, send_at
			FROM sportapp.chat_messages
			WHERE user_id = $1
			ORDER BY send_at DESC
			LIMIT $2
		) sub
		ORDER BY send_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query chat_messages: %w", err)
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			return nil, fmt.Errorf("scan chat_message: %w", err)
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows chat_messages: %w", err)
	}

	if msgs == nil {
		msgs = []ChatMessage{}
	}
	return msgs, nil
}