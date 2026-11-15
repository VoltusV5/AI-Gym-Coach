BEGIN;
CREATE TABLE sportapp.commission_settings (
 singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK(singleton),
 commission_bps INTEGER NOT NULL DEFAULT 2000 CHECK(commission_bps BETWEEN 0 AND 10000),
 effective_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 version BIGINT NOT NULL DEFAULT 1 CHECK(version>0)
);
INSERT INTO sportapp.commission_settings(singleton) VALUES(TRUE);
CREATE TABLE sportapp.commission_rates (
 version BIGINT PRIMARY KEY,
 commission_bps INTEGER NOT NULL CHECK(commission_bps BETWEEN 0 AND 10000),
 effective_at TIMESTAMPTZ NOT NULL
);
INSERT INTO sportapp.commission_rates VALUES(1,2000,'2000-01-01T00:00:00Z');
CREATE TABLE sportapp.orders (
 id UUID PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 trainer_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 program_id UUID NOT NULL REFERENCES sportapp.programs(id),
 program_version_id UUID NOT NULL REFERENCES sportapp.program_versions(id),
 status TEXT NOT NULL CHECK(status IN ('pending','paid','failed','cancelled','expired','refunded')),
 amount_minor BIGINT NOT NULL CHECK(amount_minor>0 AND amount_minor<=1000000000000),
 currency TEXT NOT NULL CHECK(currency='RUB'),
 commission_bps INTEGER NOT NULL CHECK(commission_bps BETWEEN 0 AND 10000),
 commission_minor BIGINT NOT NULL CHECK(commission_minor>=0),
 author_minor BIGINT NOT NULL CHECK(author_minor>=0),
 provider_payment_id TEXT NOT NULL UNIQUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 CHECK(amount_minor=commission_minor+author_minor)
);
CREATE UNIQUE INDEX orders_pending_unique ON sportapp.orders(user_id,program_id) WHERE status='pending';
CREATE INDEX orders_user_page ON sportapp.orders(user_id,id);
CREATE INDEX orders_trainer_page ON sportapp.orders(trainer_id,id);
CREATE TABLE sportapp.payment_events (
 id TEXT PRIMARY KEY,
 order_id UUID NOT NULL REFERENCES sportapp.orders(id),
 request_hash TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE sportapp.sales_ledger (
 order_id UUID NOT NULL REFERENCES sportapp.orders(id),
 kind TEXT NOT NULL CHECK(kind IN ('sale','refund')),
 gross_minor BIGINT NOT NULL,
 commission_minor BIGINT NOT NULL,
 author_minor BIGINT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(order_id,kind),
 CHECK(gross_minor=commission_minor+author_minor)
);
CREATE TABLE sportapp.payouts (
 id UUID PRIMARY KEY,
 trainer_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 amount_minor BIGINT NOT NULL CHECK(amount_minor>0),
 currency TEXT NOT NULL CHECK(currency='RUB'),
 status TEXT NOT NULL CHECK(status IN ('requested','processing','paid','failed','cancelled')),
 provider_payout_id TEXT,
 reason TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE sportapp.commerce_audit (
 id UUID PRIMARY KEY,
 actor_id INTEGER REFERENCES sportapp.users(id),
 action TEXT NOT NULL,
 target_id TEXT NOT NULL,
 reason TEXT NOT NULL,
 request_id TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMIT;
