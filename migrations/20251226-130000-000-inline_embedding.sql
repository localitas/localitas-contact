-- Migration: Move embedding from contact_embeddings table to inline column in contacts table

-- Add embedding column to contacts table
ALTER TABLE contacts ADD COLUMN embedding BLOB;

-- Copy existing embeddings to contacts table
UPDATE contacts SET embedding = (
    SELECT embedding FROM contact_embeddings WHERE contact_embeddings.contact_id = contacts.id
);

-- Drop the separate embeddings table
DROP TABLE IF EXISTS contact_embeddings;
