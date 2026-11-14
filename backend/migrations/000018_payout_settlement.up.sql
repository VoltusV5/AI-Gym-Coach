BEGIN;
ALTER TABLE sportapp.payouts ADD COLUMN paid_at TIMESTAMPTZ;
UPDATE sportapp.payouts SET paid_at=created_at WHERE status='paid';
ALTER TABLE sportapp.payouts ADD CONSTRAINT payout_paid_at CHECK((status='paid')=(paid_at IS NOT NULL));
COMMIT;
