CREATE TABLE IF NOT EXISTS reminders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  concept_id UUID NOT NULL,
  user_id TEXT NOT NULL,
  message TEXT NOT NULL,
  is_recurring BOOLEAN DEFAULT false,
  cron_expr TEXT,
  scheduled_at TIMESTAMPTZ,
  last_sent_at TIMESTAMPTZ,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY (concept_id) REFERENCES concepts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reminders_user_id ON reminders(user_id, is_active, scheduled_at);
