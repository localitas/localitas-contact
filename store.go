package contact

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/localitas/localitas-go"
)

const DatabaseName = "contacts"

type Store struct {
	db *sql.DB
}

// NewStore creates a Store backed by the Localitas SQL driver. The DSN should be
// the data-app URL with database_id and token query params, e.g.:
//
//	http://localhost:8090?database_id=db_123&token=base64...
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// OpenStore is a convenience that resolves the contacts database ID via the client
// SDK, then opens a database/sql connection for all subsequent operations.
func OpenStore(coreURL, dbID, token string) (*Store, error) {
	dsn := fmt.Sprintf("%s?database_id=%s&token=%s", coreURL, dbID, token)
	db, err := sql.Open("localitas", dsn)
	if err != nil {
		return nil, fmt.Errorf("open localitas db: %w", err)
	}
	return NewStore(db), nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Create(ctx context.Context, userID string, data ContactData) (*Contact, error) {
	id := newContactID()
	yaml := SerializeContactWithComments(data)
	now := time.Now().UTC().Unix()

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO contacts (id, owner_id, data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		id, userID, yaml, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert contact: %w", err)
	}

	return &Contact{
		ID:        id,
		OwnerID:   userID,
		Data:      yaml,
		CreatedAt: time.Unix(now, 0),
		UpdatedAt: time.Unix(now, 0),
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (*Contact, error) {
	var c Contact
	var createdAt, updatedAt int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id, data, created_at, updated_at FROM contacts WHERE id = ?", id,
	).Scan(&c.ID, &c.Data, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("contact %s not found", id)
	}
	c.CreatedAt = time.Unix(createdAt, 0)
	c.UpdatedAt = time.Unix(updatedAt, 0)
	return &c, nil
}

func (s *Store) List(ctx context.Context, userID string, limit, offset int) ([]*Contact, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, data, created_at, updated_at FROM contacts WHERE owner_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?",
		userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	out := make([]*Contact, 0)
	for rows.Next() {
		var c Contact
		var createdAt, updatedAt int64
		if err := rows.Scan(&c.ID, &c.Data, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = time.Unix(createdAt, 0)
		c.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &c)
	}
	return out, nil
}

func (s *Store) Update(ctx context.Context, id string, data ContactData) error {
	return s.UpdateYAML(ctx, id, SerializeContactWithComments(data))
}

func (s *Store) UpdateYAML(ctx context.Context, id, yaml string) error {
	now := time.Now().UTC().Unix()
	_, err := s.db.ExecContext(ctx,
		"UPDATE contacts SET data = ?, updated_at = ? WHERE id = ?",
		yaml, now, id)
	return err
}

func (s *Store) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM contacts WHERE id = ?", id)
	return err
}

// Search uses the contacts_fts FTS5 index created by the app's own migration.
// No dependency on the data app's global search — pure SQLite FTS.
func (s *Store) Search(ctx context.Context, userID, query string, limit int) ([]*Contact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.data, c.created_at, c.updated_at
		FROM contacts c
		JOIN contacts_fts ON c.rowid = contacts_fts.rowid
		WHERE contacts_fts MATCH ?
		AND c.owner_id = ?
		ORDER BY rank
		LIMIT ?
	`, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}
	defer rows.Close()

	out := make([]*Contact, 0)
	for rows.Next() {
		var c Contact
		var createdAt, updatedAt int64
		if err := rows.Scan(&c.ID, &c.Data, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = time.Unix(createdAt, 0)
		c.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &c)
	}
	return out, nil
}

func newContactID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
