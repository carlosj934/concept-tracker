CREATE TABLE IF NOT EXISTS concept_paths (
  ancestor_id UUID REFERENCES concepts(id) ON DELETE CASCADE,
  descendant_id UUID REFERENCES concepts(id) ON DELETE CASCADE,
  depth BIGINT,
  PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX IF NOT EXISTS idx_concept_paths_descendant ON concept_paths(descendant_id);
