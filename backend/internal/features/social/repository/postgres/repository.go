package social_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	workouts_postgres_repository "sport_app/internal/features/workouts/repository/postgres"
)

type Repository struct {
	pool         core_postgres_pool.TransactionalPool
	transactions *workouts_postgres_repository.Repository
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool, workouts_postgres_repository.NewRepository(pool)}
}
func member(ctx context.Context, r trainers_postgres_repository.Reader, user int) error {
	var allowed bool
	err := r.QueryRow(ctx, `SELECT NOT is_anonymous AND NOT blocked FROM sportapp.users WHERE id=$1`, user).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return core_errors.ErrForbidden
	}
	return nil
}
func missing(err error) error {
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return core_errors.ErrNotFound
	}
	return err
}
func available(ctx context.Context, r trainers_postgres_repository.Reader, id string) error {
	var found string
	err := r.QueryRow(ctx, `SELECT p.id::text FROM sportapp.programs p JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.id=$1 AND p.status='active' AND v.status='published'`, id).Scan(&found)
	return missing(err)
}
func eligibility(ctx context.Context, r trainers_postgres_repository.Reader, user int, program string, required int) (Eligibility, error) {
	e := Eligibility{Required: required}
	if err := member(ctx, r, user); err != nil {
		return e, err
	}
	if err := available(ctx, r, program); err != nil {
		return e, err
	}
	var entitled bool
	var author int
	err := r.QueryRow(ctx, `SELECT p.trainer_id,EXISTS(SELECT 1 FROM sportapp.entitlements WHERE program_id=p.id AND user_id=$2 AND status='active'),(SELECT COUNT(*) FROM sportapp.workout_sessions w JOIN sportapp.program_versions v ON v.id=w.source_program_version_id WHERE w.user_id=$2 AND v.program_id=p.id) FROM sportapp.programs p WHERE p.id=$1`, program, user).Scan(&author, &entitled, &e.Completed)
	if err != nil {
		return e, err
	}
	e.Eligible = entitled && e.Completed >= required && author != user
	if !e.Eligible {
		reason := "REVIEW_NOT_ELIGIBLE"
		e.Reason = &reason
	}
	return e, nil
}
func (r *Repository) Eligibility(ctx context.Context, user int, program string, required int) (Eligibility, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out, err := eligibility(ctx, r.pool, user, program, required)
	if err != nil {
		return out, err
	}
	own, err := scanReview(r.pool.QueryRow(ctx, `SELECT `+reviewColumns+` FROM sportapp.program_reviews WHERE user_id=$1 AND program_id=$2 AND deleted_at IS NULL`, user, program))
	if errors.Is(err, core_errors.ErrNotFound) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.OwnReview = &own
	return out, nil
}

const reviewColumns = `id::text,program_id::text,program_version_id::text,user_id,overall,difficulty,structure,description_match,usability,result,text,status,created_at`

func scanReview(row core_postgres_pool.Row) (Review, error) {
	var v Review
	err := row.Scan(&v.ID, &v.ProgramID, &v.ProgramVersionID, &v.UserID, &v.Overall, &v.Difficulty, &v.Structure, &v.DescriptionMatch, &v.Usability, &v.Result, &v.Text, &v.Status, &v.CreatedAt)
	return v, missing(err)
}
func (r *Repository) SaveReview(ctx context.Context, user int, program, id string, in ReviewInput, required int) (Review, error) {
	var out Review
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		if id != "" {
			if err := tx.QueryRow(ctx, `SELECT program_id::text FROM sportapp.program_reviews WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL FOR UPDATE`, id, user).Scan(&program); err != nil {
				return missing(err)
			}
		}
		eligible, err := eligibility(ctx, tx, user, program, required)
		if err != nil {
			return err
		}
		if !eligible.Eligible {
			return &core_errors.DomainError{Code: "REVIEW_NOT_ELIGIBLE", Message: "Отзыв доступен после необходимого числа тренировок", Status: 422}
		}
		if id == "" {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.program_reviews WHERE user_id=$1 AND program_id=$2 AND deleted_at IS NULL)`, user, program).Scan(&exists); err != nil {
				return err
			}
			if exists {
				return core_errors.ErrConflict
			}
			id = uuid.NewString()
			out, err = scanReview(tx.QueryRow(ctx, `INSERT INTO sportapp.program_reviews(id,user_id,program_id,program_version_id,overall,difficulty,structure,description_match,usability,result,text) SELECT $1,$2,$3,published_version_id,$4,$5,$6,$7,$8,$9,$10 FROM sportapp.programs WHERE id=$3 ON CONFLICT(user_id,program_id) DO UPDATE SET overall=EXCLUDED.overall,difficulty=EXCLUDED.difficulty,structure=EXCLUDED.structure,description_match=EXCLUDED.description_match,usability=EXCLUDED.usability,result=EXCLUDED.result,text=EXCLUDED.text,deleted_at=NULL,status='in_review' RETURNING `+reviewColumns, id, user, program, in.Overall, in.Difficulty, in.Structure, in.DescriptionMatch, in.Usability, in.Result, in.Text))
			return err
		}
		out, err = scanReview(tx.QueryRow(ctx, `UPDATE sportapp.program_reviews SET overall=$3,difficulty=$4,structure=$5,description_match=$6,usability=$7,result=$8,text=$9,status=CASE WHEN status='published' THEN 'published' ELSE 'in_review' END WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL RETURNING `+reviewColumns, id, user, in.Overall, in.Difficulty, in.Structure, in.DescriptionMatch, in.Usability, in.Result, in.Text))
		return err
	})
	return out, err
}
func (r *Repository) DeleteReview(ctx context.Context, user int, id string) error {
	return r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE sportapp.program_reviews SET deleted_at=NOW() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, user)
		if err == nil && tag.RowsAffected() == 0 {
			return core_errors.ErrNotFound
		}
		return err
	})
}
func (r *Repository) Reviews(ctx context.Context, user int, program string, admin bool, f v2request.PageFilter) (ReviewPage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := ReviewPage{Page: Page[Review]{Items: []Review{}}}
	if admin {
		if err := trainers_postgres_repository.RequireRole(ctx, r.pool, user, "admin"); err != nil {
			return out, err
		}
	} else if err := available(ctx, r.pool, program); err != nil {
		return out, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT AVG(overall)::double precision,COUNT(*) FROM sportapp.program_reviews WHERE deleted_at IS NULL AND status='published' AND ($1 OR program_id=NULLIF($2,'')::uuid)`, admin, program).Scan(&out.Average, &out.Count); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+reviewColumns+` FROM sportapp.program_reviews WHERE deleted_at IS NULL AND ($1 OR (program_id=NULLIF($2,'')::uuid AND status='published')) AND ($3='' OR id>NULLIF($3,'')::uuid) ORDER BY id LIMIT $4`, admin, program, f.After, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanReview(rows)
		if err != nil {
			return out, err
		}
		out.Items = append(out.Items, v)
	}
	return out, rows.Err()
}
func (r *Repository) ModerateReview(ctx context.Context, user int, id string, in Decision, requestID string) (Review, error) {
	var out Review
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := trainers_postgres_repository.RequireRole(ctx, tx, user, "admin"); err != nil {
			return err
		}
		var err error
		out, err = scanReview(tx.QueryRow(ctx, `UPDATE sportapp.program_reviews SET status=$2 WHERE id=$1 AND deleted_at IS NULL RETURNING `+reviewColumns, id, in.Status))
		if err != nil {
			return err
		}
		return trainers_postgres_repository.Audit(ctx, tx, user, "review_moderation", id, in.Reason, requestID)
	})
	return out, err
}

const reportColumns = `id::text,reporter_id,target_type,target_id,reason,status,created_at,resolution`

func scanReport(row core_postgres_pool.Row) (Report, error) {
	var v Report
	err := row.Scan(&v.ID, &v.ReporterID, &v.TargetType, &v.TargetID, &v.Reason, &v.Status, &v.CreatedAt, &v.Resolution)
	return v, missing(err)
}
func (r *Repository) Report(ctx context.Context, user int, in ReportInput) (Report, error) {
	var out Report
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		var exists bool
		switch in.TargetType {
		case "program":
			if err := available(ctx, tx, in.TargetID); err != nil {
				return err
			}
			exists = true
		case "trainer":
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.trainer_profiles WHERE user_id=$1::integer)`, in.TargetID).Scan(&exists); err != nil {
				return err
			}
		case "review":
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.program_reviews WHERE id=$1 AND status='published' AND deleted_at IS NULL)`, in.TargetID).Scan(&exists); err != nil {
				return err
			}
		}
		if !exists {
			return core_errors.ErrNotFound
		}
		var err error
		out, err = scanReport(tx.QueryRow(ctx, `INSERT INTO sportapp.reports(id,reporter_id,target_type,target_id,reason) VALUES($1,$2,$3,$4,$5) RETURNING `+reportColumns, uuid.NewString(), user, in.TargetType, in.TargetID, in.Reason))
		return err
	})
	return out, err
}
func (r *Repository) Reports(ctx context.Context, user int, f v2request.PageFilter) (Page[Report], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := Page[Report]{Items: []Report{}}
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, user, "admin"); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+reportColumns+` FROM sportapp.reports WHERE ($1='' OR id>NULLIF($1,'')::uuid) ORDER BY id LIMIT $2`, f.After, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanReport(rows)
		if err != nil {
			return out, err
		}
		out.Items = append(out.Items, v)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	for i := range out.Items {
		item := &out.Items[i]
		if item.TargetType != "review" {
			continue
		}
		var raw []byte
		err := r.pool.QueryRow(ctx, `SELECT to_jsonb(r) FROM sportapp.program_reviews r WHERE r.id=$1`, item.TargetID).Scan(&raw)
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			continue
		}
		if err != nil {
			return out, err
		}
		var review Review
		if err := json.Unmarshal(raw, &review); err != nil {
			return out, err
		}
		item.TargetReview = &review
	}
	return out, nil
}
func (r *Repository) DecideReport(ctx context.Context, user int, id string, in Decision, requestID string) (Report, error) {
	var out Report
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := trainers_postgres_repository.RequireRole(ctx, tx, user, "admin"); err != nil {
			return err
		}
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM sportapp.reports WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
			return missing(err)
		}
		if status == "resolved" || status == "dismissed" {
			return core_errors.ErrConflict
		}
		var err error
		out, err = scanReport(tx.QueryRow(ctx, `UPDATE sportapp.reports SET status=$2,resolution=$3 WHERE id=$1 RETURNING `+reportColumns, id, in.Status, in.Reason))
		if err != nil {
			return err
		}
		return trainers_postgres_repository.Audit(ctx, tx, user, "report_decision", id, in.Reason, requestID)
	})
	return out, err
}
func (r *Repository) Follow(ctx context.Context, user, trainer int, follow bool) error {
	return r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		if !follow {
			_, err := tx.Exec(ctx, `DELETE FROM sportapp.trainer_follows WHERE user_id=$1 AND trainer_id=$2`, user, trainer)
			return err
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.trainer_profiles t JOIN sportapp.users u ON u.id=t.user_id WHERE t.user_id=$1 AND NOT u.blocked AND EXISTS(SELECT 1 FROM sportapp.user_roles WHERE user_id=t.user_id AND role='trainer'))`, trainer).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return core_errors.ErrNotFound
		}
		_, err := tx.Exec(ctx, `INSERT INTO sportapp.trainer_follows(user_id,trainer_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, user, trainer)
		return err
	})
}
func (r *Repository) Trainers(ctx context.Context, user int, f v2request.PageFilter) (Page[trainers_postgres_repository.Profile], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := Page[trainers_postgres_repository.Profile]{Items: []trainers_postgres_repository.Profile{}}
	if err := member(ctx, r.pool, user); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+trainers_postgres_repository.ProfileJSON+` FROM sportapp.trainer_profiles t JOIN sportapp.trainer_follows f ON f.trainer_id=t.user_id JOIN sportapp.users u ON u.id=t.user_id WHERE f.user_id=$1 AND NOT u.blocked AND EXISTS(SELECT 1 FROM sportapp.user_roles ur WHERE ur.user_id=t.user_id AND ur.role='trainer') AND t.user_id>COALESCE(NULLIF($2,'')::integer,0) ORDER BY t.user_id LIMIT $3`, user, f.After, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var v trainers_postgres_repository.Profile
		if err := rows.Scan(&raw); err != nil {
			return out, err
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			return out, err
		}
		out.Items = append(out.Items, v)
	}
	return out, rows.Err()
}

const notificationColumns = `id::text,type,created_at,read_at,title,message,target_route`

func scanNotification(row core_postgres_pool.Row) (Notification, error) {
	var n Notification
	err := row.Scan(&n.ID, &n.Type, &n.CreatedAt, &n.ReadAt, &n.Title, &n.Message, &n.TargetRoute)
	return n, missing(err)
}
func (r *Repository) Notifications(ctx context.Context, user int, f v2request.PageFilter) (NotificationPage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := NotificationPage{Page: Page[Notification]{Items: []Notification{}}}
	if err := member(ctx, r.pool, user); err != nil {
		return out, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.notifications WHERE user_id=$1 AND read_at IS NULL`, user).Scan(&out.UnreadCount); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+notificationColumns+` FROM sportapp.notifications WHERE user_id=$1 AND ($2='' OR id>NULLIF($2,'')::uuid) ORDER BY id LIMIT $3`, user, f.After, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanNotification(rows)
		if err != nil {
			return out, err
		}
		out.Items = append(out.Items, v)
	}
	return out, rows.Err()
}
func (r *Repository) ReadNotification(ctx context.Context, user int, id string, read bool) (Notification, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := member(ctx, r.pool, user); err != nil {
		return Notification{}, err
	}
	return scanNotification(r.pool.QueryRow(ctx, `UPDATE sportapp.notifications SET read_at=CASE WHEN $3 THEN COALESCE(read_at,NOW()) ELSE NULL END WHERE id=$1 AND user_id=$2 RETURNING `+notificationColumns, id, user, read))
}
func (r *Repository) Events(ctx context.Context, user int, events []Event, hashes []string) (Accepted, error) {
	var out Accepted
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		for i, e := range events {
			var existing string
			err := tx.QueryRow(ctx, `SELECT request_hash FROM sportapp.behavior_events WHERE user_id=$1 AND event_id=$2`, user, e.ID).Scan(&existing)
			if err == nil {
				if existing != hashes[i] {
					return core_errors.ErrConflict
				}
				out.DuplicateCount++
				continue
			}
			if !errors.Is(err, core_postgres_pool.ErrNoRows) {
				return err
			}
			if err := available(ctx, tx, e.ProgramID); err != nil {
				return err
			}
			raw, _ := json.Marshal(e.Context)
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.behavior_events(user_id,event_id,type,program_id,occurred_at,context,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7)`, user, e.ID, e.Type, e.ProgramID, e.OccurredAt, raw, hashes[i]); err != nil {
				return err
			}
			out.Accepted++
		}
		return nil
	})
	return out, err
}
