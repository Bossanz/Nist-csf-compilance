ALTER TABLE project_subcategory_profiles
  ADD COLUMN IF NOT EXISTS assigned_user_id uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_profiles_assigned_user
  ON project_subcategory_profiles(assigned_user_id)
  WHERE assigned_user_id IS NOT NULL;
