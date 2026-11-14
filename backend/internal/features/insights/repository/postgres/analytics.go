package insights_postgres_repository

import (
	"context"
	"encoding/json"
	core_errors "sport_app/internal/core/errors"
	"time"
)

type Period struct {
	From time.Time
	To   time.Time
}

func (r *Repository) Analytics(ctx context.Context, user int, admin bool, period Period) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(c)
	}()
	if _, err = tx.Exec(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
		return nil, err
	}
	role := "trainer"
	if admin {
		role = "admin"
	}
	var authorized bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.users u JOIN sportapp.user_roles r ON r.user_id=u.id WHERE u.id=$1 AND NOT u.is_anonymous AND NOT u.blocked AND r.role=$2)`, user, role).Scan(&authorized); err != nil {
		return nil, err
	}
	if !authorized {
		return nil, core_errors.ErrForbidden
	}
	var raw []byte
	query := `WITH ledger AS (SELECT l.*,o.program_id,o.trainer_id FROM sportapp.sales_ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE l.created_at >=$2 AND l.created_at<$3 AND ($4 OR o.trainer_id=$1)),
 usage AS (SELECT DISTINCT w.user_id,v.program_id FROM sportapp.workout_sessions w JOIN sportapp.program_versions v ON v.id=w.source_program_version_id JOIN sportapp.programs p ON p.id=v.program_id WHERE w.finished_at >=$2 AND w.finished_at<$3 AND ($4 OR p.trainer_id=$1)
 UNION SELECT e.user_id,e.program_id FROM sportapp.entitlements e JOIN sportapp.programs p ON p.id=e.program_id WHERE e.granted_at >=$2 AND e.granted_at<$3 AND ($4 OR p.trainer_id=$1))
 SELECT jsonb_build_object('from',$2::timestamptz,'to',$3::timestamptz,'currency','RUB',
 'program_users',(SELECT COUNT(DISTINCT user_id) FROM usage),
 'sales_count',(SELECT COUNT(*) FROM ledger WHERE kind='sale'),
 'gross_minor',COALESCE((SELECT SUM(gross_minor) FROM ledger WHERE kind='sale'),0),
 'refund_minor',-COALESCE((SELECT SUM(gross_minor) FROM ledger WHERE kind='refund'),0),
 'commission_minor',COALESCE((SELECT SUM(commission_minor) FROM ledger),0),
 'author_earned_minor',COALESCE((SELECT SUM(author_minor) FROM ledger),0),
 'available_minor',COALESCE((SELECT SUM(l.author_minor) FROM sportapp.sales_ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.trainer_id=$1),0)-COALESCE((SELECT SUM(amount_minor) FROM sportapp.payouts WHERE trainer_id=$1 AND status IN ('requested','processing','paid')),0),
 'paid_out_minor',COALESCE((SELECT SUM(amount_minor) FROM sportapp.payouts WHERE trainer_id=$1 AND status='paid' AND paid_at >=$2 AND paid_at<$3),0),
 'programs',COALESCE((SELECT jsonb_agg(jsonb_build_object('program_id',p.id,'title',COALESCE(v.title,''),'users',(SELECT COUNT(DISTINCT u.user_id) FROM usage u WHERE u.program_id=p.id),'sales',(SELECT COUNT(*) FROM ledger l WHERE l.program_id=p.id AND l.kind='sale'),'gross_minor',COALESCE((SELECT SUM(gross_minor) FROM ledger l WHERE l.program_id=p.id AND l.kind='sale'),0)) ORDER BY p.id) FROM sportapp.programs p LEFT JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.trainer_id=$1),'[]'))`
	if admin {
		query = `WITH ledger AS (SELECT * FROM sportapp.sales_ledger WHERE created_at >=$2 AND created_at<$3),
 usage AS (SELECT DISTINCT w.user_id,v.program_id FROM sportapp.workout_sessions w JOIN sportapp.program_versions v ON v.id=w.source_program_version_id WHERE w.finished_at >=$2 AND w.finished_at<$3 UNION SELECT e.user_id,e.program_id FROM sportapp.entitlements e WHERE e.granted_at >=$2 AND e.granted_at<$3),
 active AS (SELECT user_id FROM sportapp.workout_sessions WHERE finished_at >=$2 AND finished_at<$3 UNION SELECT user_id FROM sportapp.orders o WHERE EXISTS(SELECT 1 FROM ledger l WHERE l.order_id=o.id AND l.kind='sale') UNION SELECT user_id FROM sportapp.program_reviews WHERE created_at >=$2 AND created_at<$3 AND deleted_at IS NULL)
 SELECT jsonb_build_object('from',$2::timestamptz,'to',$3::timestamptz,'currency','RUB',
 'users',(SELECT COUNT(*) FROM sportapp.users WHERE NOT is_anonymous),'active_users',(SELECT COUNT(*) FROM active),'workouts',(SELECT COUNT(*) FROM sportapp.workout_sessions WHERE finished_at >=$2 AND finished_at<$3),'published_programs',(SELECT COUNT(*) FROM sportapp.programs p JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.status='active' AND v.status='published'),
 'sales',(SELECT COUNT(*) FROM ledger WHERE kind='sale'),'gross_minor',COALESCE((SELECT SUM(gross_minor) FROM ledger WHERE kind='sale'),0),'refund_minor',-COALESCE((SELECT SUM(gross_minor) FROM ledger WHERE kind='refund'),0),'net_minor',COALESCE((SELECT SUM(gross_minor) FROM ledger),0),'commission_accrued_minor',COALESCE((SELECT SUM(commission_minor) FROM ledger),0),'commission_paid_minor',0,
 'top_programs',COALESCE((SELECT jsonb_agg(q.data ORDER BY q.sales DESC,q.id) FROM (SELECT p.id,(SELECT COUNT(*) FROM ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.program_id=p.id AND l.kind='sale') sales,jsonb_build_object('program_id',p.id,'title',COALESCE(v.title,''),'users',(SELECT COUNT(DISTINCT user_id) FROM usage u WHERE u.program_id=p.id),'sales',(SELECT COUNT(*) FROM ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.program_id=p.id AND l.kind='sale')) data FROM sportapp.programs p LEFT JOIN sportapp.program_versions v ON v.id=p.published_version_id ORDER BY sales DESC,p.id LIMIT 10) q),'[]'),
 'top_trainers',COALESCE((SELECT jsonb_agg(q.data ORDER BY q.sales DESC,q.id) FROM (SELECT t.user_id id,(SELECT COUNT(*) FROM ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.trainer_id=t.user_id AND l.kind='sale') sales,jsonb_build_object('trainer_id',t.user_id,'display_name',t.display_name,'users',(SELECT COUNT(DISTINCT u.user_id) FROM usage u JOIN sportapp.programs p ON p.id=u.program_id WHERE p.trainer_id=t.user_id),'sales',(SELECT COUNT(*) FROM ledger l JOIN sportapp.orders o ON o.id=l.order_id WHERE o.trainer_id=t.user_id AND l.kind='sale')) data FROM sportapp.trainer_profiles t ORDER BY sales DESC,t.user_id LIMIT 10) q),'[]')) WHERE $1::integer>0 AND $4::boolean`
	}
	if err = tx.QueryRow(ctx, query, user, period.From, period.To, admin).Scan(&raw); err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err = json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}
