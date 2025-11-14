package study

import "time"

// StudySession captures a single study event for a user.
type StudySession struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	SubjectID       string    `json:"subjectId"`
	Subject         string    `json:"subject"`
	Notes           string    `json:"notes"`
	Reflection      string    `json:"reflection"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationMinutes int       `json:"durationMinutes"`
	CreatedAt       time.Time `json:"createdAt"`
	LastUpdated     time.Time `json:"lastUpdated"`
}

// ProgressSummary aggregates stats for a user's study activity.
type ProgressSummary struct {
	TotalMinutes          int            `json:"totalMinutes"`
	SessionCount          int            `json:"sessionCount"`
	AverageSessionMinutes float64        `json:"averageSessionMinutes"`
	TodayMinutes          int            `json:"todayMinutes"`
	WeekMinutes           int            `json:"weekMinutes"`
	MonthMinutes          int            `json:"monthMinutes"`
	BySubject             map[string]int `json:"bySubject"`
	DailyTrend            []DailyStat    `json:"dailyTrend"`
	StreakDays            int            `json:"streakDays"`
}

// DailyStat represents aggregated stats for a single calendar day.
type DailyStat struct {
	Date           string  `json:"date"`
	TotalMinutes   int     `json:"totalMinutes"`
	SessionCount   int     `json:"sessionCount"`
	AverageMinutes float64 `json:"averageMinutes"`
}
