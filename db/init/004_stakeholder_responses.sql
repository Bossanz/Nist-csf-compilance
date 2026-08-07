CREATE TABLE IF NOT EXISTS stakeholder_responses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  subcategory_id uuid NOT NULL REFERENCES subcategories(id),
  response_text text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','submitted','reviewed','needs_more_info')),
  responded_by uuid REFERENCES users(id),
  submitted_at timestamptz,
  review_comment text NOT NULL DEFAULT '',
  reviewed_by uuid REFERENCES users(id),
  reviewed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(project_id, subcategory_id)
);
CREATE INDEX IF NOT EXISTS idx_stakeholder_responses_project ON stakeholder_responses(project_id);

CREATE TABLE IF NOT EXISTS response_documents (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  response_id uuid NOT NULL REFERENCES stakeholder_responses(id) ON DELETE CASCADE,
  original_name text NOT NULL,
  storage_key text NOT NULL UNIQUE,
  mime_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
  uploaded_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_response_documents_response ON response_documents(response_id);
