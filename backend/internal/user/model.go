package user

import "time"

// User represents an authenticated account.
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Provider     string     `json:"provider"`
	ProviderID   string     `json:"providerId"`
	IsVerified   bool       `json:"isVerified"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
