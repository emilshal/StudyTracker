package user

import (
	"context"
	"database/sql"
)

// SQLRepository persists users to SQLite.
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository constructs a user repository backed by SQLite.
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(user User) (User, error) {
	const query = `
		INSERT INTO users (id, email, password_hash, provider, provider_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(
		context.Background(),
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Provider,
		user.ProviderID,
		user.CreatedAt.UTC(),
		user.UpdatedAt.UTC(),
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) Update(user User) (User, error) {
	const query = `
		UPDATE users
		SET email = ?, password_hash = ?, provider = ?, provider_id = ?, updated_at = ?
		WHERE id = ?;
	`

	res, err := r.db.ExecContext(
		context.Background(),
		query,
		user.Email,
		user.PasswordHash,
		user.Provider,
		user.ProviderID,
		user.UpdatedAt.UTC(),
		user.ID,
	)
	if err != nil {
		return User{}, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return User{}, err
	}
	if rows == 0 {
		return User{}, sql.ErrNoRows
	}

	return user, nil
}

func (r *SQLRepository) GetByEmail(email string) (User, error) {
	const query = `
		SELECT id, email, password_hash, provider, provider_id, created_at, updated_at
		FROM users
		WHERE email = ? COLLATE NOCASE;
	`
	return r.getOne(query, email)
}

func (r *SQLRepository) GetByID(id string) (User, error) {
	const query = `
		SELECT id, email, password_hash, provider, provider_id, created_at, updated_at
		FROM users
		WHERE id = ?;
	`
	return r.getOne(query, id)
}

func (r *SQLRepository) GetByProvider(provider, providerID string) (User, error) {
	const query = `
		SELECT id, email, password_hash, provider, provider_id, created_at, updated_at
		FROM users
		WHERE provider = ? AND provider_id = ?;
	`
	return r.getOne(query, provider, providerID)
}

func (r *SQLRepository) getOne(query string, args ...interface{}) (User, error) {
	var (
		u       User
		created sql.NullTime
		updated sql.NullTime
	)
	err := r.db.QueryRowContext(context.Background(), query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Provider,
		&u.ProviderID,
		&created,
		&updated,
	)
	if err != nil {
		return User{}, err
	}

	if created.Valid {
		u.CreatedAt = created.Time.UTC()
	}
	if updated.Valid {
		u.UpdatedAt = updated.Time.UTC()
	}
	return u, nil
}
