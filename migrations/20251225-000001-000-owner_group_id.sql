-- Migration: Add owner_group_id column for group-based sharing
ALTER TABLE contacts ADD COLUMN owner_group_id TEXT;
CREATE INDEX IF NOT EXISTS idx_contacts_owner_group_id ON contacts(owner_group_id);
