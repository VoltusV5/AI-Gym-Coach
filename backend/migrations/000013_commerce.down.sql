BEGIN;
UPDATE sportapp.entitlements SET status='revoked',order_id=NULL WHERE order_id IN (SELECT id FROM sportapp.orders);
DELETE FROM sportapp.operation_results WHERE operation IN ('order','payout');
DROP TABLE sportapp.commerce_audit,sportapp.payouts,sportapp.sales_ledger,sportapp.payment_events,sportapp.orders,sportapp.commission_rates,sportapp.commission_settings;
COMMIT;
