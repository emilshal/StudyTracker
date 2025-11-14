package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Session represents a persisted login session.
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionStore manages session persistence.
type SessionStore interface {
	Create(userID string, ttl time.Duration) (Session, error)
	Get(id string) (Session, error)
	Delete(id string) error
}

// SQLSessionStore implements SessionStore using SQLite.
type SQLSessionStore struct {
	db *sql.DB
}

// NewSQLSessionStore creates a store backed by SQLite.
func NewSQLSessionStore(db *sql.DB) *SQLSessionStore {
	return &SQLSessionStore{db: db}
}

func (s *SQLSessionStore) Create(userID string, ttl time.Duration) (Session, error) {
	now := time.Now().UTC()
	session := Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
	const query = `
        INSERT INTO sessions (id, user_id, expires_at, created_at)
        VALUES (?, ?, ?, ?);
    `
	_, err := s.db.ExecContext(
		context.Background(),
		query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *SQLSessionStore) Get(id string) (Session, error) {
	const query = `
        SELECT id, user_id, expires_at, created_at
        FROM sessions
        WHERE id = ?;
    `
	var session Session
	if err := s.db.QueryRowContext(context.Background(), query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *SQLSessionStore) Delete(id string) error {
	const query = `DELETE FROM sessions WHERE id = ?;`
	_, err := s.db.ExecContext(context.Background(), query, id)
	return err
}
