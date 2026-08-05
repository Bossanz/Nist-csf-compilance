CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS functions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  description text NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  function_id uuid NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  description text NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS subcategories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  code text NOT NULL UNIQUE,
  description text NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  type text NOT NULL DEFAULT 'client' CHECK (type IN ('client','counselor_firm'))
);
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid REFERENCES organizations(id),
  name text NOT NULL,
  email text NOT NULL,
  user_type text NOT NULL CHECK (user_type IN ('counselor','stakeholder'))
);
CREATE TABLE IF NOT EXISTS projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  counselor_id uuid REFERENCES users(id),
  name text NOT NULL,
  status text NOT NULL DEFAULT 'setup' CHECK (status IN ('setup','in_review','gap_analysis','reporting','closed')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS project_functions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  function_id uuid NOT NULL REFERENCES functions(id),
  applicable boolean NOT NULL DEFAULT true,
  reason text NOT NULL DEFAULT '',
  UNIQUE(project_id, function_id)
);
CREATE TABLE IF NOT EXISTS project_subcategory_profiles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  subcategory_id uuid NOT NULL REFERENCES subcategories(id),
  included boolean NOT NULL DEFAULT false,
  rationale text NOT NULL DEFAULT '',
  current_priority text NOT NULL DEFAULT '',
  current_coverage_level text NOT NULL DEFAULT 'none' CHECK (current_coverage_level IN ('none','partial','substantial','full')),
  current_status_text text NOT NULL DEFAULT '',
  current_policies_text text NOT NULL DEFAULT '',
  current_tier text NOT NULL DEFAULT '',
  target_priority text NOT NULL DEFAULT '',
  target_coverage_level text NOT NULL DEFAULT 'none' CHECK (target_coverage_level IN ('none','partial','substantial','full')),
  target_approach_text text NOT NULL DEFAULT '',
  target_tier text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  considerations text NOT NULL DEFAULT '',
  review_status text NOT NULL DEFAULT 'draft' CHECK (review_status IN ('draft','submitted','approved','rejected')),
  submitted_at timestamptz,
  reviewed_by uuid REFERENCES users(id),
  reviewed_at timestamptz,
  UNIQUE(project_id, subcategory_id)
);
CREATE INDEX IF NOT EXISTS idx_profiles_project ON project_subcategory_profiles(project_id);
CREATE INDEX IF NOT EXISTS idx_categories_function ON categories(function_id);
CREATE INDEX IF NOT EXISTS idx_subcategories_category ON subcategories(category_id);
