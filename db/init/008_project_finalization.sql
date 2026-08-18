ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS finalized_at timestamptz,
  ADD COLUMN IF NOT EXISTS finalized_by uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_projects_finalized_at
  ON projects(finalized_at)
  WHERE finalized_at IS NOT NULL;
