-- Contacts app schema with FTS5 and vector embeddings

-- Main Contacts Table
CREATE TABLE IF NOT EXISTS contacts (
    id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contacts_updated_at ON contacts(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_contacts_created_at ON contacts(created_at DESC);

-- FTS5 Virtual Table for Keyword Search
CREATE VIRTUAL TABLE IF NOT EXISTS contacts_fts USING fts5(
    data,
    content=contacts,
    content_rowid=rowid,
    tokenize='porter unicode61'
);

CREATE TRIGGER IF NOT EXISTS contacts_ai AFTER INSERT ON contacts BEGIN
    INSERT INTO contacts_fts(rowid, data) VALUES (new.rowid, new.data);
END;

CREATE TRIGGER IF NOT EXISTS contacts_ad AFTER DELETE ON contacts BEGIN
    INSERT INTO contacts_fts(contacts_fts, rowid, data) VALUES('delete', old.rowid, old.data);
END;

CREATE TRIGGER IF NOT EXISTS contacts_au AFTER UPDATE ON contacts BEGIN
    INSERT INTO contacts_fts(contacts_fts, rowid, data) VALUES('delete', old.rowid, old.data);
    INSERT INTO contacts_fts(rowid, data) VALUES (new.rowid, new.data);
END;

-- Vector Embeddings Table for Semantic Search
CREATE TABLE IF NOT EXISTS contact_embeddings (
    contact_id TEXT PRIMARY KEY,
    embedding BLOB NOT NULL,
    dimensions INTEGER NOT NULL,
    model TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contact_embeddings_model ON contact_embeddings(model, dimensions);
CREATE INDEX IF NOT EXISTS idx_contact_embeddings_created ON contact_embeddings(created_at DESC);
