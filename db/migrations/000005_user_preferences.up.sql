CREATE TABLE IF NOT EXISTS user_preferences (
  user_id TEXT PRIMARY KEY,
  timezone TEXT DEFAULT 'UTC' NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now()
);
