BEGIN;
CREATE TABLE sportapp.public_shares (
 id UUID PRIMARY KEY,
 owner_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 token_hash TEXT NOT NULL UNIQUE,
 snapshot JSONB NOT NULL,
 source_program_version_id UUID REFERENCES sportapp.program_versions(id),
 expires_at TIMESTAMPTZ,
 revoked_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE sportapp.saved_shares (
 id UUID PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 share_id UUID NOT NULL REFERENCES sportapp.public_shares(id),
 UNIQUE(user_id,share_id)
);
COMMIT;
