CREATE TABLE concept_resources (
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

CREATE index idx_concept_resources ON concept_resources(concept_id);

CREATE index idx_concept_resources_user ON concept_resources(user_id);
