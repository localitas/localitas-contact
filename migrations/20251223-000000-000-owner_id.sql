-- Migration: Add owner_id column for UUID-based ownership
-- This allows users to change their email without breaking ownership

ALTER TABLE contacts ADD COLUMN owner_id TEXT;

CREATE INDEX IF NOT EXISTS idx_contacts_owner_id ON contacts(owner_id);
CREATE INDEX IF NOT EXISTS idx_contacts_owner_id_updated ON contacts(owner_id, updated_at DESC);
