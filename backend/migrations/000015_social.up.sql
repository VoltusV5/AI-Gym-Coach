BEGIN;
CREATE TABLE sportapp.program_reviews (
 id UUID PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 program_id UUID NOT NULL REFERENCES sportapp.programs(id),
 program_version_id UUID NOT NULL REFERENCES sportapp.program_versions(id),
 overall INTEGER NOT NULL CHECK(overall BETWEEN 1 AND 5),
 difficulty INTEGER CHECK(difficulty BETWEEN 1 AND 5),
 structure INTEGER CHECK(structure BETWEEN 1 AND 5),
 description_match INTEGER CHECK(description_match BETWEEN 1 AND 5),
 usability INTEGER CHECK(usability BETWEEN 1 AND 5),
 result INTEGER CHECK(result BETWEEN 1 AND 5),
 text TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'published' CHECK(status IN ('published','in_review','hidden')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 deleted_at TIMESTAMPTZ,
 UNIQUE(user_id,program_id)
);
CREATE INDEX reviews_program_page ON sportapp.program_reviews(program_id,id) WHERE deleted_at IS NULL;
CREATE TABLE sportapp.reports (
 id UUID PRIMARY KEY,
 reporter_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 target_type TEXT NOT NULL CHECK(target_type IN ('program','trainer','review')),
 target_id TEXT NOT NULL,
 reason TEXT NOT NULL CHECK(btrim(reason)<>''),
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','resolved','dismissed')),
 resolution TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE sportapp.trainer_follows (
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 trainer_id INTEGER NOT NULL REFERENCES sportapp.trainer_profiles(user_id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,trainer_id)
);
CREATE TABLE sportapp.notifications (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 event_key TEXT NOT NULL,
 type TEXT NOT NULL,
 title TEXT NOT NULL,
 message TEXT NOT NULL,
 target_route TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 read_at TIMESTAMPTZ,
 UNIQUE(user_id,event_key)
);
CREATE INDEX notifications_owner_page ON sportapp.notifications(user_id,id);
CREATE TABLE sportapp.behavior_events (
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 event_id UUID NOT NULL,
 type TEXT NOT NULL CHECK(type IN ('impression','view','save','unsave','recommendation_click')),
 program_id UUID NOT NULL REFERENCES sportapp.programs(id),
 occurred_at TIMESTAMPTZ NOT NULL,
 context JSONB NOT NULL,
 request_hash TEXT NOT NULL,
 PRIMARY KEY(user_id,event_id)
);
CREATE INDEX behavior_events_history ON sportapp.behavior_events(user_id,occurred_at DESC);
COMMIT;
