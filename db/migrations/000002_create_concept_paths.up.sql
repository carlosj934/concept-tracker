CREATE TABLE concept_paths (
  ancestor_id UUID REFERENCES concepts(id) ON DELETE CASCADE,
  descendant_id UUID REFERENCES concepts(id) ON DELETE CASCADE,
  depth INT,
  PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX idx_concept_paths_descendant ON concept_paths(descendant_id);
