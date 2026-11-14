package administration_service

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"sport_app/internal/core/transport/http/v2request"
	administration_postgres_repository "sport_app/internal/features/administration/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
	"strconv"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Service) References(ctx context.Context, raw, kind string, admin bool, f v2request.PageFilter, q string) (administration_postgres_repository.ReferencePage, error) {
	id := 0
	var err error
	if admin {
		id, err = v2request.UserID(raw)
	}
	if err == nil {
		err = trainers_service.ValidatePage(f, kind == "exercises")
	}
	if err != nil {
		return administration_postgres_repository.ReferencePage{}, err
	}
	if !v2request.Text(q, 200, false) {
		return administration_postgres_repository.ReferencePage{}, v2request.Invalid()
	}
	p, err := s.repo.References(ctx, id, kind, admin, f, q)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		var key struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(p.Items[f.Limit-1], &key); err != nil {
			return p, err
		}
		var after string
		if kind == "exercises" {
			after = string(key.ID)
		} else {
			if err := json.Unmarshal(key.ID, &after); err != nil {
				return p, err
			}
		}
		p.NextCursor = v2request.Next(f, after)
	}
	return p, err
}
func (s *Service) SaveReference(ctx context.Context, raw, kind, target string, in administration_postgres_repository.ReferenceInput, requestID string) (json.RawMessage, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return nil, err
	}
	if target != "" {
		if kind == "categories" {
			target, err = v2request.UUID(target)
		} else {
			n, e := strconv.Atoi(target)
			if e != nil || n < 1 || n > math.MaxInt32 {
				err = v2request.Invalid()
			}
		}
	}
	if err != nil {
		return nil, err
	}
	maxName := 200
	if kind == "exercises" {
		maxName = 100
	}
	valid := v2request.Text(in.Name, maxName, true) && v2request.Text(in.Reason, 4000, true)
	if kind == "categories" {
		valid = valid && len(in.Slug) <= 100 && slugPattern.MatchString(in.Slug) && in.EquipmentIDs == nil && in.MuscleGroup == nil && in.MuscleSubgroup == nil && in.Safe == nil && in.ImageURL == nil && in.VideoURL == nil && !in.ImageSet && !in.VideoSet && !in.SubgroupSet
		if in.Position != nil {
			valid = valid && *in.Position >= 0 && *in.Position <= math.MaxInt32
		}
	} else {
		valid = valid && in.EquipmentIDs != nil && in.Slug == "" && in.Position == nil
		for _, v := range []*string{in.MuscleGroup, in.MuscleSubgroup, in.ImageURL, in.VideoURL} {
			if v != nil {
				valid = valid && v2request.Text(*v, 100, false)
			}
		}
		if in.EquipmentIDs != nil {
			seen := map[string]bool{}
			valid = valid && len(*in.EquipmentIDs) <= 100
			for _, v := range *in.EquipmentIDs {
				valid = valid && v2request.Text(v, 200, true) && !seen[v]
				seen[v] = true
			}
		}
	}
	if !valid {
		return nil, v2request.Invalid()
	}
	return s.repo.SaveReference(ctx, id, kind, target, in, requestID)
}
