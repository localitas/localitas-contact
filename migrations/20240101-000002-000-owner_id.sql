-- Add owner_id for multi-user support

ALTER TABLE contacts ADD COLUMN owner_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_contacts_owner_id ON contacts(owner_id);
CREATE INDEX IF NOT EXISTS idx_contacts_owner_updated ON contacts(owner_id, updated_at DESC);
