ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS version_group_id uuid,
  ADD COLUMN IF NOT EXISTS version_number integer,
  ADD COLUMN IF NOT EXISTS previous_version_id uuid;

UPDATE projects
SET version_group_id = COALESCE(version_group_id, id),
    version_number = COALESCE(version_number, 1)
WHERE version_group_id IS NULL OR version_number IS NULL;

ALTER TABLE projects
  ALTER COLUMN version_group_id SET DEFAULT gen_random_uuid(),
  ALTER COLUMN version_group_id SET NOT NULL,
  ALTER COLUMN version_number SET DEFAULT 1,
  ALTER COLUMN version_number SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_version_number_check') THEN
    ALTER TABLE projects ADD CONSTRAINT projects_version_number_check CHECK (version_number > 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_previous_version_fk') THEN
    ALTER TABLE projects ADD CONSTRAINT projects_previous_version_fk
      FOREIGN KEY (previous_version_id) REFERENCES projects(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS projects_version_group_number_unique
  ON projects(version_group_id, version_number);
CREATE INDEX IF NOT EXISTS idx_projects_version_group
  ON projects(version_group_id, version_number DESC);
CREATE INDEX IF NOT EXISTS idx_projects_previous_version
  ON projects(previous_version_id);
