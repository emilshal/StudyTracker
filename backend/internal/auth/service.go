package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"studytracker/internal/user"
)

// Service coordinates auth flows.
type Service struct {
	users       user.Repository
	sessions    SessionStore
	sessionTTL  time.Duration
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

// Config contains auth configuration knobs.
type Config struct {
	SessionTTL         time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// NewService constructs an auth service.
func NewService(repo user.Repository, sessions SessionStore, cfg Config) *Service {
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 24 * time.Hour
	}
	service := &Service{
		users:      repo,
		sessions:   sessions,
		sessionTTL: cfg.SessionTTL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" && cfg.GoogleRedirectURL != "" {
		service.oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}
	return service
}

// AuthResult returns user info plus a signed session token.
type AuthResult struct {
	User    user.User `json:"user"`
	Session Session   `json:"session"`
}

// Register creates a new user using email/password credentials.
func (s *Service) Register(email, password string) (AuthResult, error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return AuthResult{}, errors.New("invalid email or password")
	}

	if _, err := s.users.GetByEmail(email); err == nil {
		return AuthResult{}, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	user := user.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Provider:     "local",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	created, err := s.users.Create(user)
	if err != nil {
		return AuthResult{}, err
	}

	session, err := s.sessions.Create(created.ID, s.sessionTTL)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: created, Session: session}, nil
}

// Login authenticates a user via email/password.
func (s *Service) Login(email, password string) (AuthResult, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return AuthResult{}, errors.New("invalid email or password")
	}

	u, err := s.users.GetByEmail(email)
	if err != nil {
		return AuthResult{}, errors.New("invalid credentials")
	}

	if u.Provider != "local" {
		return AuthResult{}, errors.New("account uses federated login")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, errors.New("invalid credentials")
	}

	session, err := s.sessions.Create(u.ID, s.sessionTTL)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: u, Session: session}, nil
}

// OAuthConfig returns the configured Google OAuth config.
func (s *Service) OAuthConfig() *oauth2.Config {
	return s.oauthConfig
}

// GetUserByID fetches a user by identifier.
func (s *Service) GetUserByID(id string) (user.User, error) {
	return s.users.GetByID(id)
}

// DeleteSession removes a persisted session.
func (s *Service) DeleteSession(sessionID string) error {
	return s.sessions.Delete(sessionID)
}

// GenerateOAuthState creates a CSRF prevention string.
func (s *Service) GenerateOAuthState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

// HandleGoogleCallback exchanges the code for tokens and signs in the user.
func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (AuthResult, error) {
	if s.oauthConfig == nil {
		return AuthResult{}, errors.New("google auth not configured")
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return AuthResult{}, fmt.Errorf("exchange failed: %w", err)
	}

	email, providerID, err := s.fetchGoogleUser(ctx, token)
	if err != nil {
		return AuthResult{}, err
	}

	u, err := s.users.GetByProvider("google", providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			now := time.Now().UTC()
			newUser := user.User{
				ID:         uuid.NewString(),
				Email:      email,
				Provider:   "google",
				ProviderID: providerID,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			u, err = s.users.Create(newUser)
			if err != nil {
				return AuthResult{}, err
			}
		} else {
			return AuthResult{}, err
		}
	}

	if u.Email == "" && email != "" {
		u.Email = email
		u.UpdatedAt = time.Now().UTC()
		if _, err := s.users.Update(u); err != nil {
			return AuthResult{}, err
		}
	}

	session, err := s.sessions.Create(u.ID, s.sessionTTL)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: u, Session: session}, nil
}

func (s *Service) fetchGoogleUser(ctx context.Context, token *oauth2.Token) (email, id string, err error) {
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("google userinfo returned %s", resp.Status)
	}
	var data struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}
	if data.ID == "" {
		return "", "", errors.New("missing google id")
	}
	return normalizeEmail(data.Email), data.ID, nil
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
