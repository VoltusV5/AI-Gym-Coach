package administration_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
)

type ReferenceInput struct {
	Reason         string    `json:"reason"`
	ImageSet       bool      `json:"-"`
	VideoSet       bool      `json:"-"`
	SubgroupSet    bool      `json:"-"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug,omitempty"`
	Position       *int      `json:"position,omitempty"`
	Active         *bool     `json:"is_active,omitempty"`
	MuscleGroup    *string   `json:"muscle_group,omitempty"`
	MuscleSubgroup *string   `json:"muscle_subgroup,omitempty"`
	EquipmentIDs   *[]string `json:"equipment_ids,omitempty"`
	Safe           *bool     `json:"safe_for_injuries,omitempty"`
	ImageURL       *string   `json:"image_url,omitempty"`
	VideoURL       *string   `json:"video_url,omitempty"`
}
type ReferencePage struct {
	Items      []json.RawMessage `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

const categoryJSON = `jsonb_build_object('id',c.id,'slug',c.slug,'name',c.name,'position',c.position,'is_active',c.is_active)`
const exerciseJSON = `jsonb_build_object('id',e.id,'name',e.exercises_name,'muscle_group',COALESCE(e.muscular_group,''),'muscle_subgroup',e.muscular_subgroup,'equipment_ids',ARRAY(SELECT equipment_id FROM sportapp.exercise_equipment WHERE exercise_id=e.id ORDER BY equipment_id),'safe_for_injuries',COALESCE(e.safe_for_injuries,FALSE),'image_url',e.image_url,'video_url',e.video_url,'is_active',e.is_active)`

func (r *Repository) References(ctx context.Context, actor int, kind string, admin bool, f v2request.PageFilter, q string) (ReferencePage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	out := ReferencePage{Items: []json.RawMessage{}}
	if admin {
		if err := trainers_postgres_repository.RequireRole(ctx, r.pool, actor, "admin"); err != nil {
			return out, err
		}
	}
	query := `SELECT ` + categoryJSON + ` FROM sportapp.program_categories c WHERE ($1 OR c.is_active) AND ($2='' OR c.id>NULLIF($2,'')::uuid) AND ($3='' OR c.name ILIKE '%' || $3 || '%') ORDER BY c.id LIMIT $4`
	if kind == "exercises" {
		query = `SELECT ` + exerciseJSON + ` FROM sportapp.exercises e WHERE ($1 OR e.is_active) AND e.id>COALESCE(NULLIF($2,'')::integer,0) AND ($3='' OR e.exercises_name ILIKE '%' || $3 || '%') ORDER BY e.id LIMIT $4`
	}
	rows, err := r.pool.Query(ctx, query, admin, f.After, q, f.Limit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return out, err
		}
		out.Items = append(out.Items, raw)
	}
	return out, rows.Err()
}
func referenceError(err error) error {
	var p *pgconn.PgError
	if errors.As(err, &p) {
		switch p.Code {
		case "23505":
			return fmt.Errorf("reference conflict: %w", core_errors.ErrConflict)
		case "23503", "23514", "22001":
			return fmt.Errorf("invalid reference: %w", core_errors.ErrInvalidArgument)
		}
	}
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return core_errors.ErrNotFound
	}
	return err
}
func (r *Repository) SaveReference(ctx context.Context, actor int, kind, id string, in ReferenceInput, requestID string) (json.RawMessage, error) {
	var out []byte
	err := r.write(ctx, actor, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if kind == "categories" {
			if id == "" {
				id = uuid.NewString()
				if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_categories(id,slug,name,position,is_active) VALUES($1,$2,$3,COALESCE($4,0),COALESCE($5,TRUE))`, id, in.Slug, in.Name, in.Position, in.Active); err != nil {
					return referenceError(err)
				}
			} else {
				tag, err := tx.Exec(ctx, `UPDATE sportapp.program_categories SET slug=$2,name=$3,position=COALESCE($4,position),is_active=COALESCE($5,is_active) WHERE id=$1`, id, in.Slug, in.Name, in.Position, in.Active)
				if err != nil {
					return referenceError(err)
				}
				if tag.RowsAffected() == 0 {
					return core_errors.ErrNotFound
				}
			}
			if err := tx.QueryRow(ctx, `SELECT `+categoryJSON+` FROM sportapp.program_categories c WHERE id=$1`, id).Scan(&out); err != nil {
				return err
			}
		} else {
			if id == "" {
				if err := tx.QueryRow(ctx, `INSERT INTO sportapp.exercises(admin_modified,exercises_name,muscular_group,muscular_subgroup,safe_for_injuries,image_url,video_url,is_active) VALUES(TRUE,$1,COALESCE($2,''),$3,COALESCE($4,FALSE),$5,$6,COALESCE($7,TRUE)) RETURNING id::text`, in.Name, in.MuscleGroup, in.MuscleSubgroup, in.Safe, in.ImageURL, in.VideoURL, in.Active).Scan(&id); err != nil {
					return referenceError(err)
				}
			} else {
				tag, err := tx.Exec(ctx, `UPDATE sportapp.exercises SET admin_modified=TRUE,exercises_name=$2,muscular_group=COALESCE($3,muscular_group),muscular_subgroup=CASE WHEN $9 THEN $4 ELSE muscular_subgroup END,safe_for_injuries=COALESCE($5,safe_for_injuries),image_url=CASE WHEN $10 THEN $6 ELSE image_url END,video_url=CASE WHEN $11 THEN $7 ELSE video_url END,is_active=COALESCE($8,is_active) WHERE id=$1::integer`, id, in.Name, in.MuscleGroup, in.MuscleSubgroup, in.Safe, in.ImageURL, in.VideoURL, in.Active, in.SubgroupSet, in.ImageSet, in.VideoSet)
				if err != nil {
					return referenceError(err)
				}
				if tag.RowsAffected() == 0 {
					return core_errors.ErrNotFound
				}
			}
			if _, err := tx.Exec(ctx, `DELETE FROM sportapp.exercise_equipment WHERE exercise_id=$1::integer`, id); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.exercise_equipment(exercise_id,equipment_id) SELECT $1::integer,unnest($2::text[])`, id, *in.EquipmentIDs); err != nil {
				return referenceError(err)
			}
			if err := tx.QueryRow(ctx, `SELECT `+exerciseJSON+` FROM sportapp.exercises e WHERE id=$1::integer`, id).Scan(&out); err != nil {
				return err
			}
		}
		target := uuid.NewString()
		if kind == "categories" {
			target = id
		}
		return trainers_postgres_repository.Audit(ctx, tx, actor, kind+":"+id, target, in.Reason, requestID)
	})
	return out, err
}
