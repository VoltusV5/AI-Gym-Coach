package trainers_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool}
}

type Reader interface {
	QueryRow(context.Context, string, ...any) core_postgres_pool.Row
}

func RequireRole(ctx context.Context, reader Reader, id int, role string) error {
	var found int
	err := reader.QueryRow(ctx, `SELECT u.id FROM sportapp.users u JOIN sportapp.user_roles r ON r.user_id=u.id
 WHERE u.id=$1 AND NOT u.is_anonymous AND r.role=$2 FOR SHARE OF u,r`, id, role).Scan(&found)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf("role required: %w", core_errors.ErrForbidden)
	}
	return err
}
func (r *Repository) Transaction(ctx context.Context, id int, role string, fn func(context.Context, core_postgres_pool.Tx) error) error {
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
	if err := RequireRole(ctx, tx, id, role); err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const profileJSON = `jsonb_build_object('id',t.user_id,'display_name',t.display_name,'photo_url',t.photo_ref,'bio',t.bio,
 'specializations',t.specializations,'experience_years',t.experience_years,'qualifications',t.qualifications,
 'verified',t.verified_at IS NOT NULL,'rating',NULL,'review_count',0,
 'published_programs',COALESCE((SELECT jsonb_agg(p.id ORDER BY p.id) FROM sportapp.programs p
 WHERE p.trainer_id=t.user_id AND p.status='active' AND p.published_version_id IS NOT NULL),'[]'::jsonb))`

func notFound(err error) error {
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf("object missing: %w", core_errors.ErrNotFound)
	}
	return err
}
func readProfile(ctx context.Context, reader Reader, id int) (Profile, error) {
	var p Profile
	var raw []byte
	err := reader.QueryRow(ctx, `SELECT `+profileJSON+` FROM sportapp.trainer_profiles t
 JOIN sportapp.users u ON u.id=t.user_id WHERE t.user_id=$1 AND NOT u.is_anonymous
 AND EXISTS(SELECT 1 FROM sportapp.user_roles r WHERE r.user_id=t.user_id AND r.role='trainer')`, id).Scan(&raw)
	if err != nil {
		return p, notFound(err)
	}
	return p, json.Unmarshal(raw, &p)
}
func (r *Repository) GetProfile(ctx context.Context, actor, id int, own bool) (Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if own {
		if err := RequireRole(ctx, r.pool, actor, "trainer"); err != nil {
			return Profile{}, err
		}
	}
	return readProfile(ctx, r.pool, id)
}
func (r *Repository) UpdateProfile(ctx context.Context, id int, p Patch) (Profile, error) {
	var result Profile
	err := r.Transaction(ctx, id, "trainer", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.trainer_profiles(user_id,display_name) VALUES($1,$2)
   ON CONFLICT(user_id) DO NOTHING`, id, p.DisplayName); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.trainer_profiles SET display_name=$2,
   photo_ref=CASE WHEN $3 THEN $4 ELSE photo_ref END,bio=COALESCE($5,bio),
   specializations=COALESCE($6,specializations),experience_years=COALESCE($7,experience_years),
   qualifications=COALESCE($8,qualifications) WHERE user_id=$1`, id, p.DisplayName, p.PhotoSet, p.PhotoRef, p.Bio, p.Specializations, p.ExperienceYears, p.Qualifications); err != nil {
			return err
		}
		var err error
		result, err = readProfile(ctx, tx, id)
		return err
	})
	return result, err
}

const verificationColumns = `id::text,trainer_id,status,qualification,document_refs,reason,created_at,reviewed_at`

func scanVerification(row core_postgres_pool.Row) (Verification, error) {
	var v Verification
	err := row.Scan(&v.ID, &v.TrainerID, &v.Status, &v.Qualification, &v.DocumentRefs, &v.Reason, &v.CreatedAt, &v.ReviewedAt)
	return v, notFound(err)
}
func (r *Repository) GetVerification(ctx context.Context, id int) (Verification, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := RequireRole(ctx, r.pool, id, "trainer"); err != nil {
		return Verification{}, err
	}
	return scanVerification(r.pool.QueryRow(ctx, `SELECT `+verificationColumns+` FROM sportapp.trainer_verifications WHERE trainer_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1`, id))
}
func (r *Repository) Submit(ctx context.Context, id int, input VerificationInput) (Verification, error) {
	var result Verification
	err := r.Transaction(ctx, id, "trainer", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		var verified bool
		if err := tx.QueryRow(ctx, `SELECT verified_at IS NOT NULL FROM sportapp.trainer_profiles WHERE user_id=$1 FOR UPDATE`, id).Scan(&verified); err != nil {
			return notFound(err)
		}
		var pending bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.trainer_verifications WHERE trainer_id=$1 AND status='pending')`, id).Scan(&pending); err != nil {
			return err
		}
		if verified || pending {
			return fmt.Errorf("verification already submitted: %w", core_errors.ErrConflict)
		}
		var err error
		result, err = scanVerification(tx.QueryRow(ctx, `INSERT INTO sportapp.trainer_verifications(id,trainer_id,qualification,document_refs)
   VALUES($1,$2,$3,$4) RETURNING `+verificationColumns, uuid.NewString(), id, input.Qualification, input.DocumentRefs))
		return err
	})
	return result, err
}
func Audit(ctx context.Context, tx core_postgres_pool.Tx, actor int, action, target, reason, requestID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO sportapp.publication_audit(id,actor_id,action,target_id,reason,request_id) VALUES($1,$2,$3,$4,$5,$6)`, uuid.NewString(), actor, action, target, reason, requestID)
	return err
}
func (r *Repository) Decide(ctx context.Context, actor int, id string, input Decision, requestID string) (Verification, error) {
	var result Verification
	err := r.Transaction(ctx, actor, "admin", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		var trainer int
		if err := tx.QueryRow(ctx, `SELECT trainer_id FROM sportapp.trainer_verifications WHERE id=$1`, id).Scan(&trainer); err != nil {
			return notFound(err)
		}
		if err := tx.QueryRow(ctx, `SELECT user_id FROM sportapp.trainer_profiles WHERE user_id=$1 FOR UPDATE`, trainer).Scan(&trainer); err != nil {
			return err
		}
		current, err := scanVerification(tx.QueryRow(ctx, `SELECT `+verificationColumns+` FROM sportapp.trainer_verifications WHERE id=$1 FOR UPDATE`, id))
		if err != nil {
			return err
		}
		if current.Status != "pending" {
			return fmt.Errorf("already decided: %w", core_errors.ErrConflict)
		}
		result, err = scanVerification(tx.QueryRow(ctx, `UPDATE sportapp.trainer_verifications SET status=$2,reason=$3,reviewed_at=NOW(),reviewed_by=$4 WHERE id=$1 RETURNING `+verificationColumns, id, input.Status, input.Reason, actor))
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.trainer_profiles SET verified_at=CASE WHEN $2='approved' THEN NOW() ELSE NULL END WHERE user_id=$1`, trainer, input.Status); err != nil {
			return err
		}
		return Audit(ctx, tx, actor, "trainer_verification", id, input.Reason, requestID)
	})
	return result, err
}
func (r *Repository) ListVerifications(ctx context.Context, actor int, f Filter) (Page[Verification], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	page := Page[Verification]{Items: []Verification{}}
	if err := RequireRole(ctx, r.pool, actor, "admin"); err != nil {
		return page, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+verificationColumns+` FROM sportapp.trainer_verifications WHERE ($1='' OR id>NULLIF($1,'')::uuid) ORDER BY id LIMIT $2`, f.After, f.Limit+1)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanVerification(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, v)
	}
	if err := rows.Err(); err != nil {
		return page, err
	}
	return page, nil
}
func (r *Repository) ListTrainers(ctx context.Context, actor int, f Filter) (Page[Profile], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	page := Page[Profile]{Items: []Profile{}}
	if err := RequireRole(ctx, r.pool, actor, "admin"); err != nil {
		return page, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+profileJSON+` FROM sportapp.trainer_profiles t WHERE t.user_id>COALESCE(NULLIF($1,'')::integer,0) ORDER BY t.user_id LIMIT $2`, f.After, f.Limit+1)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var p Profile
		if err := rows.Scan(&raw); err != nil {
			return page, err
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return page, err
		}
		page.Items = append(page.Items, p)
	}
	if err := rows.Err(); err != nil {
		return page, err
	}
	return page, nil
}
