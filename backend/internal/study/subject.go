package study

import "time"

// Subject represents a study category users can select sessions for.
type Subject struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	SessionCount int       `json:"sessionCount"`
	TotalMinutes int       `json:"totalMinutes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
