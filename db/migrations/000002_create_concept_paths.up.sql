CREATE TABLE concept_paths (
  ancestor_id UUID REFERENCES concepts(id),
  descendant_id UUID REFERENCES concepts(id),
  depth INT,
  PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX idx_concept_paths_descendant ON concept_paths(descendant_id);
