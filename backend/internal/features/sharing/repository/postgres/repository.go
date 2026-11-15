package sharing_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	workouts_postgres_repository "sport_app/internal/features/workouts/repository/postgres"
	"time"
)

type Repository struct {
	pool         core_postgres_pool.TransactionalPool
	transactions *workouts_postgres_repository.Repository
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool, workouts_postgres_repository.NewRepository(pool)}
}
func member(ctx context.Context, r trainers_postgres_repository.Reader, user int) error {
	var ok bool
	err := r.QueryRow(ctx, `SELECT NOT is_anonymous AND NOT blocked FROM sportapp.users WHERE id=$1`, user).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
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
func fresh(days []plans_postgres_repository.Day) []plans_postgres_repository.Day {
	for i := range days {
		days[i].ID = uuid.NewString()
		days[i].Position = i
	}
	return days
}
func copyPlan(ctx context.Context, tx core_postgres_pool.Tx, user int, title string, days []plans_postgres_repository.Day, version *string) (ForkResult, error) {
	if len([]rune(title)) > 200 {
		title = string([]rune(title)[:200])
	}
	p, err := plans_postgres_repository.CreateForkInTx(ctx, tx, user, plans_postgres_repository.Input{Title: title, Days: fresh(days), Schedule: plans_postgres_repository.Schedule{TimeZone: "UTC", StartDate: time.Now().UTC().Format(time.DateOnly), Entries: []plans_postgres_repository.ScheduleEntry{}}}, version)
	return ForkResult{PlanID: p.ID, SourceVersionID: version, Version: p.Version}, err
}
func program(ctx context.Context, r trainers_postgres_repository.Reader, user int, programID, version string) (PublicShare, bool, error) {
	var p PublicShare
	var entitled bool
	var price int64
	var permission bool
	err := r.QueryRow(ctx, `SELECT p.id::text,v.title,v.price_minor,v.allow_public_remix,t.user_id,t.display_name,t.verified_at IS NOT NULL,EXISTS(SELECT 1 FROM sportapp.entitlements e WHERE e.program_id=p.id AND e.user_id=$3 AND e.status='active') FROM sportapp.program_versions v JOIN sportapp.programs p ON p.id=v.program_id JOIN sportapp.trainer_profiles t ON t.user_id=p.trainer_id WHERE v.id=$2 AND ($1='' OR p.id=NULLIF($1,'')::uuid) AND v.status='published' AND p.status='active'`, programID, version, user).Scan(&p.SourceProgramID, &p.Title, &price, &permission, &p.Author.ID, &p.Author.DisplayName, &p.Author.Verified, &entitled)
	if err != nil {
		return p, false, missing(err)
	}
	p.SourceVersionID = &version
	p.RequiresAccess = price > 0 && (!entitled || !permission)
	p.CanFork = entitled
	return p, price == 0 || permission, nil
}
func programDays(ctx context.Context, reader workouts_postgres_repository.Reader, version string) ([]plans_postgres_repository.Day, error) {
	rows, err := reader.Query(ctx, `SELECT d.id::text,d.title,d.position,COALESCE((SELECT jsonb_agg(jsonb_build_object('exercise_id',e.exercise_id,'position',e.position,'sets',e.sets,'reps_min',e.reps_min,'reps_max',e.reps_max,'weight_kg',e.weight_kg,'rest_sec',e.rest_sec) ORDER BY e.position) FROM sportapp.program_exercises e WHERE e.program_day_id=d.id),'[]'::jsonb) FROM sportapp.program_days d WHERE d.program_version_id=$1 ORDER BY d.position`, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	days := []plans_postgres_repository.Day{}
	for rows.Next() {
		var d plans_postgres_repository.Day
		var raw []byte
		if err := rows.Scan(&d.ID, &d.Title, &d.Position, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &d.Exercises); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	return days, rows.Err()
}
func (r *Repository) ForkProgram(ctx context.Context, user int, id string, in ForkInput, key, hash string) (ForkResult, error) {
	var result ForkResult
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		op := "program_fork:" + id
		hit, err := workouts_postgres_repository.Cached(ctx, tx, user, op, key, hash, &result)
		if err != nil || hit {
			return err
		}
		p, _, err := program(ctx, tx, user, id, in.SourceVersionID)
		if err != nil {
			return err
		}
		if !p.CanFork {
			return core_errors.ErrForbidden
		}
		days, err := programDays(ctx, tx, in.SourceVersionID)
		if err != nil {
			return err
		}
		result, err = copyPlan(ctx, tx, user, in.Title, days, p.SourceVersionID)
		if err != nil {
			return err
		}
		return workouts_postgres_repository.SaveResult(ctx, tx, user, op, key, hash, result)
	})
	return result, err
}
func (r *Repository) Create(ctx context.Context, user int, in ShareInput, id, hash string) (ShareResult, error) {
	out := ShareResult{ID: id, ExpiresAt: in.ExpiresAt, Visibility: "public"}
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		planID := in.TargetID
		if in.TargetType == "plan_day" {
			if err := tx.QueryRow(ctx, `SELECT p.id::text FROM sportapp.plan_days d JOIN sportapp.training_plans p ON p.id=d.plan_id WHERE d.id=$1 AND p.owner_id=$2 AND p.archived_at IS NULL`, in.TargetID, user).Scan(&planID); err != nil {
				return missing(err)
			}
		}
		plan, err := plans_postgres_repository.ReadSnapshot(ctx, tx, user, planID)
		if err != nil {
			return err
		}
		snapshot := PublicShare{ID: id, Title: plan.Title, TargetType: in.TargetType, Author: Author{ID: &user, DisplayName: "Пользователь Спортика"}, Days: plan.Days, SourceVersionID: plan.SourceProgramVersionID, SnapshotAt: time.Now().UTC()}
		if in.TargetType == "plan_day" {
			snapshot.Days = []plans_postgres_repository.Day{}
			for _, d := range plan.Days {
				if d.ID == in.TargetID {
					snapshot.Days = append(snapshot.Days, d)
					snapshot.Title = d.Title
				}
			}
			if len(snapshot.Days) == 0 {
				return core_errors.ErrNotFound
			}
		}
		if snapshot.SourceVersionID != nil {
			source, permission, err := program(ctx, tx, user, "", *snapshot.SourceVersionID)
			if err != nil {
				return err
			}
			if !source.CanFork {
				return core_errors.ErrForbidden
			}
			snapshot.Author = source.Author
			snapshot.SourceProgramID = source.SourceProgramID
			if !permission {
				out.Visibility = "preview"
			}
		}
		raw, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO sportapp.public_shares(id,owner_id,token_hash,snapshot,source_program_version_id,expires_at) VALUES($1,$2,$3,$4,$5,$6)`, id, user, hash, raw, snapshot.SourceVersionID, in.ExpiresAt)
		return err
	})
	return out, err
}
func readShare(ctx context.Context, reader workouts_postgres_repository.Reader, user int, hash string) (PublicShare, error) {
	var out PublicShare
	var raw []byte
	var canRevoke bool
	err := reader.QueryRow(ctx, `SELECT snapshot,owner_id=$2 FROM sportapp.public_shares WHERE token_hash=$1 AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at>NOW()) FOR SHARE`, hash, user).Scan(&raw, &canRevoke)
	if err != nil {
		return out, missing(err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	out.CanRevoke = canRevoke
	out.CanFork = false
	if user > 0 {
		if err := reader.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.users WHERE id=$1 AND NOT is_anonymous AND NOT blocked)`, user).Scan(&out.CanFork); err != nil {
			return out, err
		}
	}
	if out.SourceVersionID != nil {
		source, permission, err := program(ctx, reader, user, "", *out.SourceVersionID)
		if err != nil {
			return out, err
		}
		out.RequiresAccess = source.RequiresAccess
		out.CanFork = source.CanFork && permission
	}
	if out.RequiresAccess {
		out.Days = []plans_postgres_repository.Day{}
	}
	return out, nil
}
func (r *Repository) Get(ctx context.Context, user int, hash string) (PublicShare, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	return readShare(ctx, r.pool, user, hash)
}
func (r *Repository) Revoke(ctx context.Context, user int, id string) error {
	return r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE sportapp.public_shares SET revoked_at=COALESCE(revoked_at,NOW()) WHERE id=$1 AND owner_id=$2`, id, user)
		if err == nil && tag.RowsAffected() == 0 {
			return core_errors.ErrNotFound
		}
		return err
	})
}
func (r *Repository) Save(ctx context.Context, user int, hash string) (string, error) {
	var id string
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		p, err := readShare(ctx, tx, user, hash)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `INSERT INTO sportapp.saved_shares(id,user_id,share_id) VALUES($1,$2,$3) ON CONFLICT(user_id,share_id) DO UPDATE SET share_id=EXCLUDED.share_id RETURNING id::text`, uuid.NewString(), user, p.ID).Scan(&id)
	})
	return id, err
}
func (r *Repository) ForkShare(ctx context.Context, user int, hash, key, payloadHash string) (ForkResult, error) {
	var out ForkResult
	err := r.transactions.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		op := "share_fork:" + hash
		hit, err := workouts_postgres_repository.Cached(ctx, tx, user, op, key, payloadHash, &out)
		if err != nil || hit {
			return err
		}
		p, err := readShare(ctx, tx, user, hash)
		if err != nil {
			return err
		}
		if !p.CanFork || p.RequiresAccess {
			return core_errors.ErrForbidden
		}
		out, err = copyPlan(ctx, tx, user, p.Title+" — моя версия", p.Days, p.SourceVersionID)
		if err != nil {
			return err
		}
		return workouts_postgres_repository.SaveResult(ctx, tx, user, op, key, payloadHash, out)
	})
	return out, err
}
