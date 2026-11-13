package commerce_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	workouts_postgres_repository "sport_app/internal/features/workouts/repository/postgres"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository { return &Repository{pool} }
func (r *Repository) write(ctx context.Context, fn func(context.Context, core_postgres_pool.Tx) error) error {
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
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(784321)`); err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func member(ctx context.Context, r trainers_postgres_repository.Reader, id int) error {
	var ok bool
	if err := r.QueryRow(ctx, `SELECT NOT is_anonymous AND NOT blocked FROM sportapp.users WHERE id=$1 FOR SHARE`, id).Scan(&ok); err != nil {
		return missing(err)
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

const orderColumns = `id::text,status,amount_minor,currency,program_id::text,program_version_id::text,trainer_id,user_id,commission_bps,commission_minor,author_minor,provider_payment_id,created_at`

func scanOrder(row core_postgres_pool.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.Status, &o.AmountMinor, &o.Currency, &o.ProgramID, &o.ProgramVersionID, &o.TrainerID, &o.UserID, &o.CommissionBps, &o.CommissionMinor, &o.AuthorMinor, &o.PaymentID, &o.CreatedAt)
	return o, missing(err)
}
func readOrder(ctx context.Context, r trainers_postgres_repository.Reader, id string) (Order, error) {
	return scanOrder(r.QueryRow(ctx, `SELECT `+orderColumns+` FROM sportapp.orders WHERE id=$1`, id))
}
func (r *Repository) Create(ctx context.Context, user int, in OrderInput, key, hash string) (Order, error) {
	var out Order
	err := r.write(ctx, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := member(ctx, tx, user); err != nil {
			return err
		}
		hit, err := workouts_postgres_repository.Cached(ctx, tx, user, "order", key, hash, &out)
		if err != nil || hit {
			return err
		}
		var owned bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.entitlements WHERE user_id=$1 AND program_id=$2 AND status='active')`, user, in.ProgramID).Scan(&owned); err != nil {
			return err
		}
		if owned {
			return &core_errors.DomainError{Code: "ALREADY_OWNED", Message: "Программа уже доступна", Status: 409}
		}
		var price, bps int64
		var currency string
		var trainer int
		if err := tx.QueryRow(ctx, `SELECT v.price_minor,v.currency,p.trainer_id FROM sportapp.programs p JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.id=$1 AND v.id=$2 AND p.status='active' AND v.status='published' FOR SHARE OF p,v`, in.ProgramID, in.ProgramVersionID).Scan(&price, &currency, &trainer); err != nil {
			return missing(err)
		}
		if price <= 0 || price > 1000000000000 || currency != "RUB" {
			return core_errors.ErrUnprocessable
		}
		var pending bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.orders WHERE user_id=$1 AND program_id=$2 AND status='pending')`, user, in.ProgramID).Scan(&pending); err != nil {
			return err
		}
		if pending {
			return core_errors.ErrConflict
		}
		if err := tx.QueryRow(ctx, `SELECT commission_bps FROM sportapp.commission_rates WHERE effective_at<=NOW() ORDER BY effective_at DESC,version DESC LIMIT 1`).Scan(&bps); err != nil {
			return err
		}
		commission := (price*bps + 5000) / 10000
		id := uuid.NewString()
		out, err = scanOrder(tx.QueryRow(ctx, `INSERT INTO sportapp.orders(id,user_id,trainer_id,program_id,program_version_id,status,amount_minor,currency,commission_bps,commission_minor,author_minor,provider_payment_id) VALUES($1,$2,$3,$4,$5,'pending',$6,$7,$8,$9,$10,$11) RETURNING `+orderColumns, id, user, trainer, in.ProgramID, in.ProgramVersionID, price, currency, bps, commission, price-commission, "stub-"+id))
		if err != nil {
			return err
		}
		return workouts_postgres_repository.SaveResult(ctx, tx, user, "order", key, hash, out)
	})
	return out, err
}
func (r *Repository) Get(ctx context.Context, user int, id string) (Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := member(ctx, r.pool, user); err != nil {
		return Order{}, err
	}
	o, err := readOrder(ctx, r.pool, id)
	if err == nil && o.UserID != user {
		return Order{}, core_errors.ErrNotFound
	}
	return o, err
}
func (r *Repository) List(ctx context.Context, user int, admin bool, f v2request.PageFilter) (Page[Order], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	p := Page[Order]{Items: []Order{}}
	if admin {
		if err := trainers_postgres_repository.RequireRole(ctx, r.pool, user, "admin"); err != nil {
			return p, err
		}
	} else if err := member(ctx, r.pool, user); err != nil {
		return p, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+orderColumns+` FROM sportapp.orders WHERE ($1 OR user_id=$2) AND ($3='' OR id>NULLIF($3,'')::uuid) ORDER BY id LIMIT $4`, admin, user, f.After, f.Limit+1)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return p, err
		}
		p.Items = append(p.Items, o)
	}
	return p, rows.Err()
}
func audit(ctx context.Context, tx core_postgres_pool.Tx, actor *int, action, target, reason, requestID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO sportapp.commerce_audit(id,actor_id,action,target_id,reason,request_id) VALUES($1,$2,$3,$4,$5,$6)`, uuid.NewString(), actor, action, target, reason, requestID)
	return err
}
func (r *Repository) Confirm(ctx context.Context, e Event, hash, requestID string) error {
	return r.write(ctx, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		var previous string
		err := tx.QueryRow(ctx, `SELECT request_hash FROM sportapp.payment_events WHERE id=$1`, e.EventID).Scan(&previous)
		if err == nil {
			if previous != hash {
				return core_errors.ErrConflict
			}
			return nil
		}
		if !errors.Is(err, core_postgres_pool.ErrNoRows) {
			return err
		}
		o, err := readOrder(ctx, tx, e.OrderID)
		if err != nil {
			return err
		}
		if o.PaymentID != e.PaymentID || o.AmountMinor != e.AmountMinor || o.Currency != e.Currency {
			return core_errors.ErrInvalidArgument
		}
		status := map[string]string{"payment_succeeded": "paid", "payment_failed": "failed", "refund_succeeded": "refunded"}[e.Type]
		if status == "" {
			return core_errors.ErrInvalidArgument
		}
		if o.Status != status {
			validTransition := (o.Status == "pending" && (status == "paid" || status == "failed")) || (o.Status == "paid" && status == "refunded")
			if !validTransition {
				return core_errors.ErrConflict
			}
			if _, err := tx.Exec(ctx, `UPDATE sportapp.orders SET status=$2 WHERE id=$1`, o.ID, status); err != nil {
				return err
			}
			if status == "paid" {
				if _, err := tx.Exec(ctx, `INSERT INTO sportapp.entitlements(id,user_id,program_id,status,order_id) VALUES($1,$2,$3,'active',$4) ON CONFLICT(user_id,program_id) DO UPDATE SET status='active',granted_at=NOW(),order_id=EXCLUDED.order_id`, uuid.NewString(), o.UserID, o.ProgramID, o.ID); err != nil {
					return err
				}
			}
			if status == "refunded" {
				if _, err := tx.Exec(ctx, `UPDATE sportapp.entitlements SET status='revoked' WHERE order_id=$1`, o.ID); err != nil {
					return err
				}
			}
			if status == "paid" || status == "refunded" {
				kind, sign := "sale", int64(1)
				if status == "refunded" {
					kind, sign = "refund", -1
				}
				if _, err := tx.Exec(ctx, `INSERT INTO sportapp.sales_ledger(order_id,kind,gross_minor,commission_minor,author_minor) VALUES($1,$2,$3,$4,$5)`, o.ID, kind, o.AmountMinor*sign, o.CommissionMinor*sign, o.AuthorMinor*sign); err != nil {
					return err
				}
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.payment_events(id,order_id,request_hash) VALUES($1,$2,$3)`, e.EventID, o.ID, hash); err != nil {
			return err
		}
		return audit(ctx, tx, nil, e.Type, o.ID, "verified stub event", requestID)
	})
}
func (r *Repository) Commission(ctx context.Context, actor int) (Commission, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var c Commission
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, actor, "admin"); err != nil {
		return c, err
	}
	err := r.pool.QueryRow(ctx, `SELECT commission_bps,effective_at,version FROM sportapp.commission_settings WHERE singleton`).Scan(&c.Bps, &c.EffectiveAt, &c.Version)
	return c, err
}
func (r *Repository) SetCommission(ctx context.Context, actor int, c Commission, requestID string) (Commission, error) {
	var out Commission
	err := r.write(ctx, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := trainers_postgres_repository.RequireRole(ctx, tx, actor, "admin"); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `UPDATE sportapp.commission_settings SET commission_bps=$1,effective_at=$2,version=version+1 WHERE singleton AND version=$3 RETURNING commission_bps,effective_at,version`, c.Bps, c.EffectiveAt, c.Version).Scan(&out.Bps, &out.EffectiveAt, &out.Version)
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_errors.ErrConflict
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.commission_rates(version,commission_bps,effective_at) VALUES($1,$2,$3)`, out.Version, out.Bps, out.EffectiveAt); err != nil {
			return err
		}
		return audit(ctx, tx, &actor, "commission", "settings", fmt.Sprint(c.Bps), requestID)
	})
	return out, err
}

const payoutColumns = `id::text,trainer_id,amount_minor,currency,status,created_at,provider_payout_id`

func scanPayout(row core_postgres_pool.Row) (Payout, error) {
	var p Payout
	err := row.Scan(&p.ID, &p.TrainerID, &p.AmountMinor, &p.Currency, &p.Status, &p.CreatedAt, &p.ProviderPayoutID)
	return p, missing(err)
}
func (r *Repository) Payouts(ctx context.Context, actor int, admin bool, f v2request.PageFilter) (Page[Payout], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	p := Page[Payout]{Items: []Payout{}}
	role := "trainer"
	if admin {
		role = "admin"
	}
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, actor, role); err != nil {
		return p, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+payoutColumns+` FROM sportapp.payouts WHERE ($1 OR trainer_id=$2) AND ($3='' OR id>NULLIF($3,'')::uuid) ORDER BY id LIMIT $4`, admin, actor, f.After, f.Limit+1)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanPayout(rows)
		if err != nil {
			return p, err
		}
		p.Items = append(p.Items, v)
	}
	return p, rows.Err()
}
func (r *Repository) CreatePayout(ctx context.Context, actor int, in PayoutInput, key, hash, requestID string) (Payout, error) {
	var out Payout
	err := r.write(ctx, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := trainers_postgres_repository.RequireRole(ctx, tx, actor, "admin"); err != nil {
			return err
		}
		hit, err := workouts_postgres_repository.Cached(ctx, tx, actor, "payout", key, hash, &out)
		if err != nil || hit {
			return err
		}
		if err := trainers_postgres_repository.RequireRole(ctx, tx, in.TrainerID, "trainer"); err != nil {
			return err
		}
		var available int64
		if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT SUM(l.author_minor) FROM sportapp.sales_ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.trainer_id=$1 AND o.currency=$2),0)-COALESCE((SELECT SUM(amount_minor) FROM sportapp.payouts WHERE trainer_id=$1 AND currency=$2 AND status IN ('requested','processing','paid')),0)`, in.TrainerID, in.Currency).Scan(&available); err != nil {
			return err
		}
		if available < in.AmountMinor {
			return core_errors.ErrUnprocessable
		}
		id := uuid.NewString()
		out, err = scanPayout(tx.QueryRow(ctx, `INSERT INTO sportapp.payouts(id,trainer_id,amount_minor,currency,status,provider_payout_id) VALUES($1,$2,$3,$4,'paid',$5) RETURNING `+payoutColumns, id, in.TrainerID, in.AmountMinor, in.Currency, "stub-payout-"+id))
		if err != nil {
			return err
		}
		if err := audit(ctx, tx, &actor, "stub_payout", id, "test only", requestID); err != nil {
			return err
		}
		return workouts_postgres_repository.SaveResult(ctx, tx, actor, "payout", key, hash, out)
	})
	return out, err
}
func (r *Repository) DecidePayout(ctx context.Context, actor int, id string, in Decision, requestID string) (Payout, error) {
	var out Payout
	err := r.write(ctx, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		if err := trainers_postgres_repository.RequireRole(ctx, tx, actor, "admin"); err != nil {
			return err
		}
		p, err := scanPayout(tx.QueryRow(ctx, `SELECT `+payoutColumns+` FROM sportapp.payouts WHERE id=$1 FOR UPDATE`, id))
		if err != nil {
			return err
		}
		if in.Action == "reconcile" {
			out = p
			return audit(ctx, tx, &actor, "payout_reconcile", id, in.Reason, requestID)
		}
		if p.Status == "paid" || p.Status == "processing" || p.Status == "cancelled" {
			return core_errors.ErrConflict
		}
		status := "cancelled"
		if in.Action == "retry" {
			var available int64
			if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT SUM(l.author_minor) FROM sportapp.sales_ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.trainer_id=$1 AND o.currency=$2),0)-COALESCE((SELECT SUM(amount_minor) FROM sportapp.payouts WHERE trainer_id=$1 AND currency=$2 AND status IN ('requested','processing','paid') AND id<>$3),0)`, p.TrainerID, p.Currency, p.ID).Scan(&available); err != nil {
				return err
			}
			if available < p.AmountMinor {
				return core_errors.ErrUnprocessable
			}
			status = "paid"
		}
		out, err = scanPayout(tx.QueryRow(ctx, `UPDATE sportapp.payouts SET status=$2,reason=$3,provider_payout_id=CASE WHEN $2='paid' THEN 'stub-payout-' || id::text ELSE provider_payout_id END WHERE id=$1 RETURNING `+payoutColumns, id, status, in.Reason))
		if err != nil {
			return err
		}
		return audit(ctx, tx, &actor, "payout_"+in.Action, id, in.Reason, requestID)
	})
	return out, err
}
