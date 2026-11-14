BEGIN;
ALTER TABLE sportapp.payouts DROP CONSTRAINT payout_paid_at;
ALTER TABLE sportapp.payouts DROP COLUMN paid_at;
COMMIT;
