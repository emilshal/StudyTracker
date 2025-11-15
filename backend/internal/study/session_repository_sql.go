package study

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"studytracker/internal/platform/database"
)

// SQLSessionRepository persists study sessions to SQLite.
type SQLSessionRepository struct {
	db        *sql.DB
	useDollar bool
}

// NewSQLSessionRepository returns a SessionRepository backed by SQLite.
func NewSQLSessionRepository(db *sql.DB) *SQLSessionRepository {
	return &SQLSessionRepository{
		db:        db,
		useDollar: database.UsesDollarPlaceholders(db),
	}
}

func (r *SQLSessionRepository) Create(session StudySession) (StudySession, error) {
	const query = `
		INSERT INTO study_sessions (
			id, user_id, subject_id, subject_name, notes, reflection,
			start_time, end_time, duration_minutes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err := r.db.ExecContext(
		context.Background(),
		r.rebind(query),
		session.ID,
		session.UserID,
		session.SubjectID,
		session.Subject,
		nullIfEmpty(session.Notes),
		nullIfEmpty(session.Reflection),
		session.StartTime.UTC(),
		session.EndTime.UTC(),
		session.DurationMinutes,
		session.CreatedAt.UTC(),
		session.LastUpdated.UTC(),
	)
	if err != nil {
		return StudySession{}, err
	}

	return session, nil
}

func (r *SQLSessionRepository) Update(session StudySession) (StudySession, error) {
	const query = `
		UPDATE study_sessions
		SET subject_id = ?, subject_name = ?, notes = ?, reflection = ?,
			start_time = ?, end_time = ?, duration_minutes = ?, updated_at = ?
		WHERE id = ? AND user_id = ?;
	`

	res, err := r.db.ExecContext(
		context.Background(),
		r.rebind(query),
		session.SubjectID,
		session.Subject,
		nullIfEmpty(session.Notes),
		nullIfEmpty(session.Reflection),
		session.StartTime.UTC(),
		session.EndTime.UTC(),
		session.DurationMinutes,
		session.LastUpdated.UTC(),
		session.ID,
		session.UserID,
	)
	if err != nil {
		return StudySession{}, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return StudySession{}, err
	}
	if rows == 0 {
		return StudySession{}, ErrNotFound
	}

	return session, nil
}

func (r *SQLSessionRepository) Delete(userID, id string) error {
	const query = `DELETE FROM study_sessions WHERE id = ? AND user_id = ?;`

	res, err := r.db.ExecContext(context.Background(), r.rebind(query), id, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLSessionRepository) List(userID string) ([]StudySession, error) {
	const query = `
		SELECT id, user_id, subject_id, subject_name, notes, reflection,
		       start_time, end_time, duration_minutes, created_at, updated_at
		FROM study_sessions
		WHERE user_id = ?
		ORDER BY start_time DESC;
	`

	rows, err := r.db.QueryContext(context.Background(), r.rebind(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []StudySession
	for rows.Next() {
		var session StudySession
		var notes sql.NullString
		var reflection sql.NullString
		var start, end, created, updated time.Time

		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SubjectID,
			&session.Subject,
			&notes,
			&reflection,
			&start,
			&end,
			&session.DurationMinutes,
			&created,
			&updated,
		); err != nil {
			return nil, err
		}

		if notes.Valid {
			session.Notes = notes.String
		}
		if reflection.Valid {
			session.Reflection = reflection.String
		}
		session.StartTime = start.UTC()
		session.EndTime = end.UTC()
		session.CreatedAt = created.UTC()
		session.LastUpdated = updated.UTC()

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SQLSessionRepository) rebind(query string) string {
	return database.Rebind(query, r.useDollar)
}

func nullIfEmpty(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: value,
		Valid:  true,
	}
}
