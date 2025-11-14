package study

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"strings"
	"time"
)

const (
	dayLayout = "2006-01-02"
	trendDays = 14
)

// Service contains the business logic for study tracking.
type Service struct {
	sessions SessionRepository
	subjects SubjectRepository
}

// NewService constructs a service with the provided repositories.
func NewService(sessionRepo SessionRepository, subjectRepo SubjectRepository) *Service {
	return &Service{
		sessions: sessionRepo,
		subjects: subjectRepo,
	}
}

// Study session operations ----------------------------------------------------

// CreateSession stores a new study session, generating an ID when missing.
func (s *Service) CreateSession(userID string, session StudySession) (StudySession, error) {
	session.UserID = userID
	if err := s.prepareSession(&session, true); err != nil {
		return StudySession{}, err
	}

	return s.sessions.Create(session)
}

// UpdateSession persists changes to an existing study session.
func (s *Service) UpdateSession(userID string, session StudySession) (StudySession, error) {
	session.UserID = userID
	if err := s.prepareSession(&session, false); err != nil {
		return StudySession{}, err
	}

	return s.sessions.Update(session)
}

// DeleteSession removes a study session.
func (s *Service) DeleteSession(userID, id string) error {
	return s.sessions.Delete(userID, id)
}

// ListSessions returns all stored sessions; extend with filters later.
func (s *Service) ListSessions(userID string) ([]StudySession, error) {
	return s.sessions.List(userID)
}

// BuildSummary aggregates study data for dashboards.
func (s *Service) BuildSummary(userID string) (ProgressSummary, error) {
	sessions, err := s.sessions.List(userID)
	if err != nil {
		return ProgressSummary{}, err
	}

	summary := ProgressSummary{
		BySubject: make(map[string]int),
	}

	if len(sessions) == 0 {
		summary.DailyTrend = buildEmptyTrend()
		return summary, nil
	}

	loc := time.Local
	now := time.Now().In(loc)
	startToday := startOfDay(now)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := startToday.AddDate(0, 0, -(weekday - 1))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	dailyMap := make(map[string]*DailyStat)

	for _, session := range sessions {
		summary.TotalMinutes += session.DurationMinutes
		summary.SessionCount++
		summary.BySubject[session.Subject] += session.DurationMinutes

		start := session.StartTime.In(loc)
		day := startOfDay(start)
		key := day.Format(dayLayout)

		stat, ok := dailyMap[key]
		if !ok {
			stat = &DailyStat{Date: key}
			dailyMap[key] = stat
		}
		stat.TotalMinutes += session.DurationMinutes
		stat.SessionCount++
		stat.AverageMinutes = float64(stat.TotalMinutes) / float64(stat.SessionCount)

		if day.Equal(startToday) {
			summary.TodayMinutes += session.DurationMinutes
		}
		if !day.Before(weekStart) {
			summary.WeekMinutes += session.DurationMinutes
		}
		if !day.Before(monthStart) {
			summary.MonthMinutes += session.DurationMinutes
		}
	}

	if summary.SessionCount > 0 {
		summary.AverageSessionMinutes = float64(summary.TotalMinutes) / float64(summary.SessionCount)
	}

	summary.DailyTrend = buildDailyTrend(startToday, dailyMap)
	summary.StreakDays = calculateStreak(startToday, dailyMap)

	return summary, nil
}

// Subject operations ----------------------------------------------------------

// ListSubjects returns subjects in alphabetical order.
func (s *Service) ListSubjects(userID string) ([]Subject, error) {
	if s.subjects == nil {
		return nil, nil
	}
	return s.subjects.List(userID)
}

// CreateSubject adds a new subject to the catalogue.
func (s *Service) CreateSubject(userID string, subject Subject) (Subject, error) {
	if s.subjects == nil {
		return Subject{}, errors.New("subject repository not configured")
	}

	subject.UserID = userID
	subject.Name = strings.TrimSpace(subject.Name)
	subject.Color = strings.TrimSpace(subject.Color)
	if subject.Name == "" {
		return Subject{}, ErrSubjectNameEmpty
	}

	now := time.Now().UTC()
	if subject.ID == "" {
		subject.ID = generateID()
	}
	subject.CreatedAt = now
	subject.UpdatedAt = now

	return s.subjects.Create(subject)
}

// UpdateSubject allows renaming or recolouring a subject.
func (s *Service) UpdateSubject(userID string, subject Subject) (Subject, error) {
	if s.subjects == nil {
		return Subject{}, errors.New("subject repository not configured")
	}

	subject.UserID = userID
	subject.Name = strings.TrimSpace(subject.Name)
	subject.Color = strings.TrimSpace(subject.Color)
	if subject.Name == "" {
		return Subject{}, ErrSubjectNameEmpty
	}

	existing, err := s.subjects.Get(userID, subject.ID)
	if err != nil {
		return Subject{}, err
	}

	subject.CreatedAt = existing.CreatedAt
	subject.UpdatedAt = time.Now().UTC()

	return s.subjects.Update(subject)
}

// DeleteSubject removes a subject from the catalogue.
func (s *Service) DeleteSubject(userID, id string) error {
	if s.subjects == nil {
		return errors.New("subject repository not configured")
	}
	return s.subjects.Delete(userID, id)
}

// prepareSession validates and normalises session data prior to persistence.
func (s *Service) prepareSession(session *StudySession, isCreate bool) error {
	session.Subject = strings.TrimSpace(session.Subject)
	if session.Subject == "" {
		return ErrMissingSubject
	}

	if s.subjects != nil && session.UserID != "" {
		subject, err := s.subjects.GetByName(session.UserID, session.Subject)
		if err != nil {
			if errors.Is(err, ErrSubjectNotFound) {
				return ErrUnknownSubject
			}
			return err
		}
		session.SubjectID = subject.ID
		session.Subject = subject.Name
	}

	if session.StartTime.IsZero() || session.EndTime.IsZero() {
		return ErrInvalidTiming
	}

	session.StartTime = session.StartTime.UTC()
	session.EndTime = session.EndTime.UTC()

	if !session.EndTime.After(session.StartTime) {
		return ErrInvalidTiming
	}

	duration := session.EndTime.Sub(session.StartTime).Minutes()
	session.DurationMinutes = int(math.Ceil(duration))

	now := time.Now().UTC()
	session.LastUpdated = now
	if isCreate {
		if session.ID == "" {
			session.ID = generateID()
		}
		session.CreatedAt = now
	}

	return nil
}

func buildEmptyTrend() []DailyStat {
	now := time.Now().In(time.Local)
	return buildDailyTrend(startOfDay(now), map[string]*DailyStat{})
}

func buildDailyTrend(startToday time.Time, daily map[string]*DailyStat) []DailyStat {
	trend := make([]DailyStat, 0, trendDays)
	for i := trendDays - 1; i >= 0; i-- {
		day := startToday.AddDate(0, 0, -i)
		key := day.Format(dayLayout)
		if stat, ok := daily[key]; ok {
			trend = append(trend, *stat)
		} else {
			trend = append(trend, DailyStat{
				Date:           key,
				TotalMinutes:   0,
				SessionCount:   0,
				AverageMinutes: 0,
			})
		}
	}
	return trend
}

func calculateStreak(startToday time.Time, daily map[string]*DailyStat) int {
	streak := 0
	cursor := startToday
	for {
		key := cursor.Format(dayLayout)
		stat, ok := daily[key]
		if !ok || stat.TotalMinutes == 0 {
			break
		}
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func generateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
