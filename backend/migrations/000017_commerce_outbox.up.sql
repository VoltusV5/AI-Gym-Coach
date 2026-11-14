BEGIN;
CREATE TABLE sportapp.commerce_outbox (
 event_key TEXT PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 type TEXT NOT NULL,
 title TEXT NOT NULL,
 message TEXT NOT NULL,
 target_route TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 delivered_at TIMESTAMPTZ
);
CREATE INDEX commerce_outbox_pending ON sportapp.commerce_outbox(created_at,event_key) WHERE delivered_at IS NULL;
COMMIT;
