CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (lower(email));

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'invited';
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();

UPDATE users
SET role = CASE WHEN user_type = 'counselor' THEN 'counselor' ELSE 'viewer' END
WHERE role IS NULL;

UPDATE users SET status = 'active' WHERE password_hash IS NOT NULL AND status = 'invited';

ALTER TABLE users ALTER COLUMN role SET NOT NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check') THEN
    ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_status_check') THEN
    ALTER TABLE users ADD CONSTRAINT users_status_check CHECK (status IN ('invited','active','disabled'));
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_organization_ownership_check') THEN
    ALTER TABLE users ADD CONSTRAINT users_organization_ownership_check CHECK (
      (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
      (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer'))
    );
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS invitations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid REFERENCES organizations(id) ON DELETE CASCADE,
  email text NOT NULL,
  user_type text NOT NULL CHECK (user_type IN ('counselor','stakeholder')),
  role text NOT NULL CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer')),
  token_hash text NOT NULL UNIQUE,
  invited_by uuid NOT NULL REFERENCES users(id),
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT invitations_scope_check CHECK (
    (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
    (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer'))
  )
);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(lower(email));
CREATE INDEX IF NOT EXISTS idx_invitations_expires ON invitations(expires_at);

CREATE TABLE IF NOT EXISTS audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id uuid REFERENCES users(id),
  organization_id uuid REFERENCES organizations(id),
  project_id uuid REFERENCES projects(id),
  action text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_organization ON audit_logs(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_project ON audit_logs(project_id, created_at DESC);
