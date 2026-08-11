ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS objective text NOT NULL DEFAULT '';

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS assessment_period text NOT NULL DEFAULT '';

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS target_completion_date date;

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS scope_boundary text NOT NULL DEFAULT '';

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS compliance_driver text NOT NULL DEFAULT '';
