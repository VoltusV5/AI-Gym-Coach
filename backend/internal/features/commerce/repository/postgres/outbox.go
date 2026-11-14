package commerce_postgres_repository

import (
	"context"
)

func (r *Repository) DispatchNotifications(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(c)
	}()
	_, err = tx.Exec(ctx, `WITH batch AS MATERIALIZED (SELECT * FROM sportapp.commerce_outbox WHERE delivered_at IS NULL ORDER BY created_at,event_key FOR UPDATE SKIP LOCKED LIMIT 100),
 inserted AS (INSERT INTO sportapp.notifications(user_id,event_key,type,title,message,target_route) SELECT user_id,event_key,type,title,message,target_route FROM batch ON CONFLICT(user_id,event_key) DO NOTHING RETURNING id)
 UPDATE sportapp.commerce_outbox SET delivered_at=NOW() WHERE event_key IN (SELECT event_key FROM batch)`)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
