package study

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"studytracker/internal/platform/database"
)

// SQLSubjectRepository persists subjects to SQLite.
type SQLSubjectRepository struct {
	db        *sql.DB
	useDollar bool
}

// NewSQLSubjectRepository returns a SubjectRepository backed by SQLite.
func NewSQLSubjectRepository(db *sql.DB) *SQLSubjectRepository {
	return &SQLSubjectRepository{
		db:        db,
		useDollar: database.UsesDollarPlaceholders(db),
	}
}

func (r *SQLSubjectRepository) Create(subject Subject) (Subject, error) {
	const query = `
		INSERT INTO subjects (id, user_id, name, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?);
	`

	_, err := r.db.ExecContext(
		context.Background(),
		r.rebind(query),
		subject.ID,
		subject.UserID,
		subject.Name,
		nullIfEmpty(subject.Color),
		subject.CreatedAt.UTC(),
		subject.UpdatedAt.UTC(),
	)
	if err != nil {
		return Subject{}, mapSubjectError(err)
	}

	return subject, nil
}

func (r *SQLSubjectRepository) Update(subject Subject) (Subject, error) {
	const query = `
		UPDATE subjects
		SET name = ?, color = ?, updated_at = ?
		WHERE id = ? AND user_id = ?;
	`

	res, err := r.db.ExecContext(
		context.Background(),
		r.rebind(query),
		subject.Name,
		nullIfEmpty(subject.Color),
		subject.UpdatedAt.UTC(),
		subject.ID,
		subject.UserID,
	)
	if err != nil {
		return Subject{}, mapSubjectError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return Subject{}, err
	}
	if rows == 0 {
		return Subject{}, ErrSubjectNotFound
	}

	return subject, nil
}

func (r *SQLSubjectRepository) Delete(userID, id string) error {
	const query = `DELETE FROM subjects WHERE id = ? AND user_id = ?;`

	res, err := r.db.ExecContext(context.Background(), r.rebind(query), id, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSubjectNotFound
	}

	return nil
}

func (r *SQLSubjectRepository) List(userID string) ([]Subject, error) {
	const query = `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM subjects
		WHERE user_id = ?
		ORDER BY LOWER(name) ASC;
	`

	rows, err := r.db.QueryContext(context.Background(), r.rebind(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []Subject
	for rows.Next() {
		var subject Subject
		var color sql.NullString
		var created, updated time.Time

		if err := rows.Scan(
			&subject.ID,
			&subject.UserID,
			&subject.Name,
			&color,
			&created,
			&updated,
		); err != nil {
			return nil, err
		}

		if color.Valid {
			subject.Color = color.String
		}
		subject.CreatedAt = created.UTC()
		subject.UpdatedAt = updated.UTC()

		subjects = append(subjects, subject)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subjects, nil
}

func (r *SQLSubjectRepository) Get(userID, id string) (Subject, error) {
	const query = `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM subjects
		WHERE id = ? AND user_id = ?;
	`

	var subject Subject
	var color sql.NullString
	var created, updated time.Time
	err := r.db.QueryRowContext(context.Background(), r.rebind(query), id, userID).Scan(
		&subject.ID,
		&subject.UserID,
		&subject.Name,
		&color,
		&created,
		&updated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subject{}, ErrSubjectNotFound
		}
		return Subject{}, err
	}

	if color.Valid {
		subject.Color = color.String
	}
	subject.CreatedAt = created.UTC()
	subject.UpdatedAt = updated.UTC()

	return subject, nil
}

func (r *SQLSubjectRepository) GetByName(userID, name string) (Subject, error) {
	const query = `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM subjects
		WHERE user_id = ? AND LOWER(name) = LOWER(?);
	`

	var subject Subject
	var color sql.NullString
	var created, updated time.Time
	err := r.db.QueryRowContext(context.Background(), r.rebind(query), userID, name).Scan(
		&subject.ID,
		&subject.UserID,
		&subject.Name,
		&color,
		&created,
		&updated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subject{}, ErrSubjectNotFound
		}
		return Subject{}, err
	}

	if color.Valid {
		subject.Color = color.String
	}
	subject.CreatedAt = created.UTC()
	subject.UpdatedAt = updated.UTC()

	return subject, nil
}

func mapSubjectError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "UNIQUE constraint failed: subjects.name") {
		return ErrSubjectNameExists
	}
	if strings.Contains(err.Error(), "duplicate key value") && strings.Contains(err.Error(), "subjects_name") {
		return ErrSubjectNameExists
	}
	return err
}

func (r *SQLSubjectRepository) rebind(query string) string {
	return database.Rebind(query, r.useDollar)
}
