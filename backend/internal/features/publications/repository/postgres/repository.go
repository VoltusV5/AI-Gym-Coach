package publications_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
)

type Repository struct {
	pool         core_postgres_pool.TransactionalPool
	transactions *trainers_postgres_repository.Repository
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool, transactions: trainers_postgres_repository.NewRepository(pool)}
}

const programJSON = `jsonb_build_object('id',p.id,'version',p.version,'published_version_id',p.published_version_id,
 'version_id',v.id,'status',CASE WHEN p.status='hidden' THEN 'hidden' ELSE v.status END,
 'draft',jsonb_build_object('version',p.version,'title',v.title,'description',v.description,'goal',v.goal,'level',v.level,
 'duration_weeks',v.duration_weeks,'sessions_per_week',v.sessions_per_week,'price_minor',v.price_minor,'currency',v.currency,
 'allow_public_remix',v.allow_public_remix,
 'equipment_ids',COALESCE((SELECT jsonb_agg(e.equipment_id ORDER BY e.equipment_id) FROM sportapp.program_version_equipment e WHERE e.version_id=v.id),'[]'::jsonb),
 'category_ids',COALESCE((SELECT jsonb_agg(c.category_id ORDER BY c.category_id) FROM sportapp.program_version_categories c WHERE c.version_id=v.id),'[]'::jsonb),
 'days',COALESCE((SELECT jsonb_agg(jsonb_build_object('id',d.id,'title',d.title,'position',d.position,
 'exercises',COALESCE((SELECT jsonb_agg(jsonb_build_object('exercise_id',ex.exercise_id,'position',ex.position,'sets',ex.sets,'reps_min',ex.reps_min,'reps_max',ex.reps_max,'weight_kg',ex.weight_kg,'rest_sec',ex.rest_sec) ORDER BY ex.position)
 FROM sportapp.program_exercises ex WHERE ex.program_day_id=d.id),'[]'::jsonb)) ORDER BY d.position)
 FROM sportapp.program_days d WHERE d.program_version_id=v.id),'[]'::jsonb)))`
const programFrom = ` FROM sportapp.programs p JOIN LATERAL (SELECT * FROM sportapp.program_versions WHERE program_id=p.id ORDER BY number DESC LIMIT 1) v ON TRUE `

func missing(err error) error {
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf("program missing: %w", core_errors.ErrNotFound)
	}
	return err
}
func read(ctx context.Context, reader trainers_postgres_repository.Reader, owner int, id string) (Program, error) {
	var p Program
	var raw []byte
	err := reader.QueryRow(ctx, `SELECT `+programJSON+programFrom+` WHERE p.id=$1 AND ($2=0 OR p.trainer_id=$2)`, id, owner).Scan(&raw)
	if err != nil {
		return p, missing(err)
	}
	return p, json.Unmarshal(raw, &p)
}
func (r *Repository) Get(ctx context.Context, owner int, id string) (Program, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, owner, "trainer"); err != nil {
		return Program{}, err
	}
	return read(ctx, r.pool, owner, id)
}
func (r *Repository) List(ctx context.Context, actor int, admin bool, f Filter) (Page, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	page := Page{Items: []Program{}}
	role := "trainer"
	owner := actor
	if admin {
		role = "admin"
		owner = 0
	}
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, actor, role); err != nil {
		return page, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+programJSON+programFrom+` WHERE ($1=0 OR p.trainer_id=$1) AND ($2='' OR p.id>NULLIF($2,'')::uuid) ORDER BY p.id LIMIT $3`, owner, f.After, f.Limit+1)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var p Program
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
func referenceError(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && (pg.Code == "23503" || pg.Code == "23505" || pg.Code == "23514") {
		return fmt.Errorf("invalid program references: %w", core_errors.ErrInvalidArgument)
	}
	return err
}
func saveDraft(ctx context.Context, tx core_postgres_pool.Tx, programID, versionID string, number int, d Draft, create bool) error {
	var valid bool
	ids := []int{}
	dayIDs := []string{}
	for _, day := range d.Days {
		dayIDs = append(dayIDs, day.ID)
		for _, ex := range day.Exercises {
			ids = append(ids, ex.ExerciseID)
		}
	}
	if err := tx.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM unnest($1::integer[]) x(id) WHERE NOT EXISTS(SELECT 1 FROM sportapp.exercises e WHERE e.id=x.id AND e.is_active))
 AND NOT EXISTS(SELECT 1 FROM unnest($2::text[]) x(id) WHERE NOT EXISTS(SELECT 1 FROM sportapp.equipment e WHERE e.id=x.id))
 AND NOT EXISTS(SELECT 1 FROM unnest($3::uuid[]) x(id) WHERE NOT EXISTS(SELECT 1 FROM sportapp.program_categories c WHERE c.id=x.id AND c.is_active))
 AND NOT EXISTS(SELECT 1 FROM sportapp.program_days WHERE id=ANY($4::uuid[]) AND program_version_id<>$5)`, ids, d.EquipmentIDs, d.CategoryIDs, dayIDs, versionID).Scan(&valid); err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid program references: %w", core_errors.ErrInvalidArgument)
	}
	if create {
		_, err := tx.Exec(ctx, `INSERT INTO sportapp.program_versions(id,program_id,number,title,description,goal,level,duration_weeks,sessions_per_week,price_minor,currency,allow_public_remix,status)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'draft')`, versionID, programID, number, d.Title, d.Description, d.Goal, d.Level, d.DurationWeeks, d.SessionsPerWeek, d.PriceMinor, d.Currency, d.AllowPublicRemix)
		if err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx, `UPDATE sportapp.program_versions SET title=$2,description=$3,goal=$4,level=$5,duration_weeks=$6,sessions_per_week=$7,price_minor=$8,currency=$9,allow_public_remix=$10,status='draft' WHERE id=$1 AND status IN ('draft','rejected')`, versionID, d.Title, d.Description, d.Goal, d.Level, d.DurationWeeks, d.SessionsPerWeek, d.PriceMinor, d.Currency, d.AllowPublicRemix); err != nil {
			return err
		}
	}
	for _, q := range []string{`DELETE FROM sportapp.program_exercises WHERE program_day_id IN(SELECT id FROM sportapp.program_days WHERE program_version_id=$1)`, `DELETE FROM sportapp.program_days WHERE program_version_id=$1`, `DELETE FROM sportapp.program_version_equipment WHERE version_id=$1`, `DELETE FROM sportapp.program_version_categories WHERE version_id=$1`} {
		if _, err := tx.Exec(ctx, q, versionID); err != nil {
			return err
		}
	}
	for _, id := range d.EquipmentIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_version_equipment(version_id,equipment_id) VALUES($1,$2)`, versionID, id); err != nil {
			return err
		}
	}
	for _, id := range d.CategoryIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_version_categories(version_id,category_id) VALUES($1,$2)`, versionID, id); err != nil {
			return err
		}
	}
	for _, day := range d.Days {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_days(id,program_version_id,title,position) VALUES($1,$2,$3,$4)`, day.ID, versionID, day.Title, day.Position); err != nil {
			return err
		}
		for _, ex := range day.Exercises {
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_exercises(id,program_day_id,exercise_id,position,sets,reps_min,reps_max,weight_kg,rest_sec) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.NewString(), day.ID, ex.ExerciseID, ex.Position, ex.Sets, ex.RepsMin, ex.RepsMax, ex.WeightKg, ex.RestSec); err != nil {
				return err
			}
		}
	}
	return nil
}
func lock(ctx context.Context, tx core_postgres_pool.Tx, owner int, id string) error {
	var found string
	return missing(tx.QueryRow(ctx, `SELECT id::text FROM sportapp.programs WHERE id=$1 AND ($2=0 OR trainer_id=$2) FOR UPDATE`, id, owner).Scan(&found))
}
func (r *Repository) Save(ctx context.Context, owner int, id string, d Draft) (Program, error) {
	var result Program
	err := r.transactions.Transaction(ctx, owner, "trainer", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if id == "" {
			id = uuid.NewString()
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.programs(id,trainer_id) VALUES($1,$2)`, id, owner); err != nil {
				return err
			}
			if err := saveDraft(ctx, tx, id, uuid.NewString(), 1, d, true); err != nil {
				return err
			}
		} else {
			if err := lock(ctx, tx, owner, id); err != nil {
				return err
			}
			current, err := read(ctx, tx, owner, id)
			if err != nil {
				return err
			}
			if current.Version != d.Version {
				return fmt.Errorf("program changed: %w", core_errors.ErrConflict)
			}
			if current.Status != "draft" && current.Status != "rejected" {
				return fmt.Errorf("version immutable: %w", core_errors.ErrConflict)
			}
			if err := saveDraft(ctx, tx, id, current.VersionID, 0, d, false); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE sportapp.programs SET version=version+1 WHERE id=$1`, id); err != nil {
				return err
			}
		}
		var err error
		result, err = read(ctx, tx, owner, id)
		return err
	})
	return result, referenceError(err)
}
func (r *Repository) NewVersion(ctx context.Context, owner int, id string) (Program, error) {
	var result Program
	err := r.transactions.Transaction(ctx, owner, "trainer", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := lock(ctx, tx, owner, id); err != nil {
			return err
		}
		current, err := read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if current.Status != "published" {
			return fmt.Errorf("existing draft: %w", core_errors.ErrConflict)
		}
		var number int
		if err := tx.QueryRow(ctx, `SELECT number+1 FROM sportapp.program_versions WHERE id=$1`, current.VersionID).Scan(&number); err != nil {
			return err
		}
		d := current.Draft
		for i := range d.Days {
			d.Days[i].ID = uuid.NewString()
		}
		if err := saveDraft(ctx, tx, id, uuid.NewString(), number, d, true); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.programs SET version=version+1 WHERE id=$1`, id); err != nil {
			return err
		}
		result, err = read(ctx, tx, owner, id)
		return err
	})
	return result, referenceError(err)
}
func publishable(ctx context.Context, tx core_postgres_pool.Tx, id string) error {
	var valid bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.program_days WHERE program_version_id=$1)
 AND NOT EXISTS(SELECT 1 FROM sportapp.program_days d WHERE d.program_version_id=$1 AND NOT EXISTS(SELECT 1 FROM sportapp.program_exercises ex WHERE ex.program_day_id=d.id))
 AND NOT EXISTS(SELECT 1 FROM sportapp.program_exercises ex JOIN sportapp.program_days d ON d.id=ex.program_day_id JOIN sportapp.exercises e ON e.id=ex.exercise_id WHERE d.program_version_id=$1 AND (NOT e.is_active OR EXISTS(SELECT 1 FROM sportapp.exercise_equipment eq WHERE eq.exercise_id=e.id AND NOT EXISTS(SELECT 1 FROM sportapp.program_version_equipment pe WHERE pe.version_id=$1 AND pe.equipment_id=eq.equipment_id))))`, id).Scan(&valid); err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("program incomplete or incompatible: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
func (r *Repository) Publish(ctx context.Context, owner int, id string, version int64) (Workflow, error) {
	var out Workflow
	err := r.transactions.Transaction(ctx, owner, "trainer", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := lock(ctx, tx, owner, id); err != nil {
			return err
		}
		p, err := read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if p.Version != version || (p.Status != "draft" && p.Status != "rejected") {
			return fmt.Errorf("cannot submit version: %w", core_errors.ErrConflict)
		}
		if err := publishable(ctx, tx, p.VersionID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.program_versions SET status='in_review' WHERE id=$1`, p.VersionID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.programs SET version=version+1 WHERE id=$1`, id); err != nil {
			return err
		}
		out = Workflow{ID: id, Status: "in_review", Version: p.Version + 1}
		return nil
	})
	return out, err
}
func (r *Repository) Moderate(ctx context.Context, actor int, id string, m Moderation, requestID string) (Workflow, error) {
	var out Workflow
	err := r.transactions.Transaction(ctx, actor, "admin", func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := lock(ctx, tx, 0, id); err != nil {
			return err
		}
		p, err := read(ctx, tx, 0, id)
		if err != nil {
			return err
		}
		if p.Status != "in_review" || (m.ProgramVersionID != "" && p.VersionID != m.ProgramVersionID) {
			return fmt.Errorf("version no longer in review: %w", core_errors.ErrConflict)
		}
		status := "rejected"
		if m.Status == "approved" || m.Status == "published" {
			status = "published"
			if err := publishable(ctx, tx, p.VersionID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.program_versions SET status=$2 WHERE id=$1`, p.VersionID, status); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.programs SET version=version+1,published_version_id=CASE WHEN $2='published' THEN $3::uuid ELSE published_version_id END,
   status=CASE WHEN $2='published' THEN 'active' ELSE status END WHERE id=$1`, id, status, p.VersionID); err != nil {
			return err
		}
		if err := trainers_postgres_repository.Audit(ctx, tx, actor, "program_moderation", p.VersionID, m.Reason, requestID); err != nil {
			return err
		}
		out = Workflow{ID: id, Status: status, Version: p.Version + 1}
		return nil
	})
	return out, err
}
