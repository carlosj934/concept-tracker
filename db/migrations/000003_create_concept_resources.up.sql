CREATE TABLE IF NOT EXISTS concept_resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  concept_id UUID NOT NULL,
  user_id TEXT NOT NULL,
  provider TEXT NOT NULL,
  external_id TEXT NOT NULL,
  url TEXT NOT NULL,
  title TEXT,
  meta JSONB,
  created_at TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY (concept_id) REFERENCES concepts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_concept_resources ON concept_resources(concept_id);

CREATE INDEX IF NOT EXISTS idx_concept_resources_user ON concept_resources(user_id);
