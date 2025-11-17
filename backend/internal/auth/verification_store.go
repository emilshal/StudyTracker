package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/google/uuid"

	"studytracker/internal/platform/database"
)

// VerificationToken represents a short-lived token emailed to the user.
type VerificationToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// VerificationTokenStore persists verification tokens.
type VerificationTokenStore interface {
	Create(userID string, ttl time.Duration) (VerificationToken, error)
	GetByToken(token string) (VerificationToken, error)
	Delete(id string) error
	DeleteByUser(userID string) error
}

// SQLVerificationTokenStore implements VerificationTokenStore backed by SQL.
type SQLVerificationTokenStore struct {
	db        *sql.DB
	useDollar bool
}

// NewSQLVerificationTokenStore constructs a SQL-backed token store.
func NewSQLVerificationTokenStore(db *sql.DB) *SQLVerificationTokenStore {
	return &SQLVerificationTokenStore{
		db:        db,
		useDollar: database.UsesDollarPlaceholders(db),
	}
}

func (s *SQLVerificationTokenStore) Create(userID string, ttl time.Duration) (VerificationToken, error) {
	now := time.Now().UTC()
	tokenValue, err := generateVerificationToken()
	if err != nil {
		return VerificationToken{}, err
	}

	token := VerificationToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     tokenValue,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	const query = `
        INSERT INTO verification_tokens (id, user_id, token, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?);
    `

	if _, err := s.db.ExecContext(
		context.Background(),
		s.rebind(query),
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	); err != nil {
		return VerificationToken{}, err
	}

	return token, nil
}

func (s *SQLVerificationTokenStore) GetByToken(token string) (VerificationToken, error) {
	const query = `
        SELECT id, user_id, token, expires_at, created_at
        FROM verification_tokens
        WHERE token = ?;
    `

	var vt VerificationToken
	if err := s.db.QueryRowContext(context.Background(), s.rebind(query), token).Scan(
		&vt.ID,
		&vt.UserID,
		&vt.Token,
		&vt.ExpiresAt,
		&vt.CreatedAt,
	); err != nil {
		return VerificationToken{}, err
	}
	return vt, nil
}

func (s *SQLVerificationTokenStore) Delete(id string) error {
	const query = `DELETE FROM verification_tokens WHERE id = ?;`
	_, err := s.db.ExecContext(context.Background(), s.rebind(query), id)
	return err
}

func (s *SQLVerificationTokenStore) DeleteByUser(userID string) error {
	const query = `DELETE FROM verification_tokens WHERE user_id = ?;`
	_, err := s.db.ExecContext(context.Background(), s.rebind(query), userID)
	return err
}

func (s *SQLVerificationTokenStore) rebind(query string) string {
	return database.Rebind(query, s.useDollar)
}

func generateVerificationToken() (string, error) {
	const size = 32
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
