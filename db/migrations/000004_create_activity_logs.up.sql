CREATE TABLE IF NOT EXISTS activity_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  concept_id UUID NOT NULL,
  user_id TEXT NOT NULL,
  activity_type TEXT NOT NULL,
  duration_minutes INT,
  notes TEXT,
  logged_at TIMESTAMPTZ DEFAULT now(),
  created_at TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY (concept_id) REFERENCES concepts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_concept_activity_logs ON activity_logs(concept_id);

CREATE INDEX IF NOT EXISTS idx_concept_activity_logs_user ON activity_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_concept_logged_at ON activity_logs(logged_at DESC);
