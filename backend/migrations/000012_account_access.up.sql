BEGIN;
ALTER TABLE sportapp.users ADD COLUMN blocked BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE sportapp.users ADD COLUMN block_reason TEXT NOT NULL DEFAULT '';
CREATE TABLE sportapp.account_audit (
 id UUID PRIMARY KEY,
 actor_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 target_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 reason TEXT NOT NULL CHECK (btrim(reason) <> ''),
 request_id TEXT NOT NULL,
 before_state JSONB NOT NULL,
 after_state JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO sportapp.user_roles(user_id,role) SELECT id,'user' FROM sportapp.users WHERE NOT is_anonymous ON CONFLICT DO NOTHING;
CREATE FUNCTION sportapp.ensure_user_role() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NOT NEW.is_anonymous THEN
  INSERT INTO sportapp.user_roles(user_id,role) VALUES(NEW.id,'user') ON CONFLICT DO NOTHING;
 END IF;
 RETURN NEW;
END;
$$;
CREATE TRIGGER ensure_user_role AFTER INSERT OR UPDATE OF is_anonymous ON sportapp.users FOR EACH ROW EXECUTE FUNCTION sportapp.ensure_user_role();
COMMIT;
