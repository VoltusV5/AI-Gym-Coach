BEGIN;
ALTER TABLE sportapp.trainer_profiles ADD COLUMN specializations TEXT[] NOT NULL DEFAULT '{}';
CREATE TABLE sportapp.trainer_verifications (
 id UUID PRIMARY KEY,
 trainer_id INTEGER NOT NULL REFERENCES sportapp.trainer_profiles(user_id),
 status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
 qualification TEXT NOT NULL CHECK (btrim(qualification) <> ''),
 document_refs TEXT[] NOT NULL,
 reason TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ,
 reviewed_by INTEGER REFERENCES sportapp.users(id),
 CHECK ((status = 'pending' AND reviewed_at IS NULL AND reviewed_by IS NULL) OR
        (status <> 'pending' AND reviewed_at IS NOT NULL AND reviewed_by IS NOT NULL AND btrim(reason) <> ''))
);
CREATE UNIQUE INDEX trainer_verifications_pending_uniq ON sportapp.trainer_verifications(trainer_id) WHERE status = 'pending';
CREATE INDEX trainer_verifications_owner_idx ON sportapp.trainer_verifications(trainer_id,created_at DESC,id DESC);
CREATE TABLE sportapp.publication_audit (
 id UUID PRIMARY KEY,
 actor_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 action TEXT NOT NULL,
 target_id UUID NOT NULL,
 reason TEXT NOT NULL,
 request_id TEXT NOT NULL,
 occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE sportapp.programs ADD COLUMN version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0);
COMMIT;
