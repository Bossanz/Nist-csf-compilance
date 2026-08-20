ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
  CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer','auditor'));

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_organization_ownership_check;
ALTER TABLE users ADD CONSTRAINT users_organization_ownership_check CHECK (
  (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
  (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer','auditor'))
);

ALTER TABLE invitations ADD COLUMN IF NOT EXISTS cancelled_at timestamptz;
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS cancelled_by uuid REFERENCES users(id);
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS superseded_at timestamptz;
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS superseded_by uuid REFERENCES invitations(id);
ALTER TABLE invitations DROP CONSTRAINT IF EXISTS invitations_role_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_role_check
  CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer','auditor'));
ALTER TABLE invitations DROP CONSTRAINT IF EXISTS invitations_scope_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_scope_check CHECK (
  (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
  (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer','auditor'))
);

CREATE TABLE IF NOT EXISTS invitation_project_access (
  invitation_id uuid NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  PRIMARY KEY (invitation_id, project_id)
);

CREATE TABLE IF NOT EXISTS project_auditor_access (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  granted_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  revoked_at timestamptz,
  PRIMARY KEY (project_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_project_auditor_access_user
  ON project_auditor_access(user_id, revoked_at);

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_role text;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS result text NOT NULL DEFAULT 'success';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id text;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS ip_address inet;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_agent text;
