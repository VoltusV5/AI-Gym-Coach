package sharing_postgres_repository

import (
	"context"
	core_errors "sport_app/internal/core/errors"
	"sport_app/internal/core/transport/http/v2request"
)

type SavedItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	TargetType string `json:"target_type"`
	Available  bool   `json:"available"`
}
type SavedPage struct {
	Items      []SavedItem `json:"items"`
	NextCursor *string     `json:"next_cursor"`
}

func (r *Repository) Saved(ctx context.Context, user int, f v2request.PageFilter) (SavedPage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := SavedPage{Items: []SavedItem{}}
	if err := member(ctx, r.pool, user); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT s.id::text,p.snapshot->>'title',p.snapshot->>'target_type',p.revoked_at IS NULL AND (p.expires_at IS NULL OR p.expires_at>NOW()) FROM sportapp.saved_shares s JOIN sportapp.public_shares p ON p.id=s.share_id WHERE s.user_id=$1 AND ($2='' OR s.id>NULLIF($2,'')::uuid) ORDER BY s.id LIMIT $3`, user, f.After, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var item SavedItem
		if err := rows.Scan(&item.ID, &item.Title, &item.TargetType, &item.Available); err != nil {
			return out, err
		}
		out.Items = append(out.Items, item)
	}
	return out, rows.Err()
}
func (r *Repository) SavedHash(ctx context.Context, user int, id string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := member(ctx, r.pool, user); err != nil {
		return "", err
	}
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT p.token_hash FROM sportapp.saved_shares s JOIN sportapp.public_shares p ON p.id=s.share_id WHERE s.user_id=$1 AND s.id=$2`, user, id).Scan(&hash)
	return hash, missing(err)
}
func (r *Repository) RemoveSaved(ctx context.Context, user int, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tag, err := r.pool.Exec(ctx, `DELETE FROM sportapp.saved_shares WHERE user_id=$1 AND id=$2`, user, id)
	if err == nil && tag.RowsAffected() == 0 {
		return core_errors.ErrNotFound
	}
	return err
}
