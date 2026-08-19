CREATE TABLE IF NOT EXISTS remediation_actions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  subcategory_id uuid NOT NULL REFERENCES subcategories(id),
  title text NOT NULL CHECK (btrim(title) <> ''),
  description text NOT NULL DEFAULT '',
  desired_result text NOT NULL DEFAULT '',
  priority text NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'critical')),
  owner_user_id uuid NOT NULL REFERENCES users(id),
  due_date date NOT NULL,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'awaiting_review', 'closed')),
  progress_note text NOT NULL DEFAULT '',
  review_comment text NOT NULL DEFAULT '',
  created_by uuid NOT NULL REFERENCES users(id),
  submitted_at timestamptz,
  closed_by uuid REFERENCES users(id),
  closed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS remediation_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  action_id uuid NOT NULL REFERENCES remediation_actions(id) ON DELETE CASCADE,
  original_name text NOT NULL,
  storage_path text NOT NULL,
  mime_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
  uploaded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_remediation_actions_project_status
  ON remediation_actions(project_id, status);

CREATE INDEX IF NOT EXISTS idx_remediation_actions_owner_status
  ON remediation_actions(owner_user_id, status);

CREATE INDEX IF NOT EXISTS idx_remediation_actions_outcome
  ON remediation_actions(project_id, subcategory_id);

CREATE INDEX IF NOT EXISTS idx_remediation_evidence_action
  ON remediation_evidence(action_id, created_at, id);
