-- Migration: Add contacts_members table for resource-level authorization

CREATE TABLE IF NOT EXISTS contacts_members (
    contact_id TEXT NOT NULL,
    user_id TEXT NOT NULL DEFAULT '',
    group_id TEXT NOT NULL DEFAULT '',
    permission TEXT NOT NULL DEFAULT 'read',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    PRIMARY KEY (contact_id, user_id, group_id),
    CHECK (permission IN ('read', 'write', 'admin')),
    CHECK ((user_id != '' AND group_id = '') OR (user_id = '' AND group_id != '')),
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contacts_members_user_id ON contacts_members(user_id);
CREATE INDEX IF NOT EXISTS idx_contacts_members_group_id ON contacts_members(group_id);
CREATE INDEX IF NOT EXISTS idx_contacts_members_contact_id ON contacts_members(contact_id);
