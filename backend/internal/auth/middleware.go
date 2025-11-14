package auth

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Middleware validates auth cookies and injects the user ID into the context.
type Middleware struct {
	sessions  SessionStore
	cookieKey string
}

// NewMiddleware constructs an auth middleware.
func NewMiddleware(store SessionStore) *Middleware {
	return &Middleware{
		sessions:  store,
		cookieKey: "session_token",
	}
}

// RequireAuth ensures the request includes a valid session cookie.
func (m *Middleware) RequireAuth(c *fiber.Ctx) error {
	sessionID := c.Cookies(m.cookieKey)
	if sessionID == "" {
		logUnauthorized(c, "missing session cookie")
		return fiber.ErrUnauthorized
	}

	session, err := m.sessions.Get(sessionID)
	if err != nil {
		logUnauthorized(c, "session lookup failed")
		return fiber.ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) {
		_ = m.sessions.Delete(sessionID)
		logUnauthorized(c, "session expired")
		return fiber.ErrUnauthorized
	}

	c.Locals(ContextUserIDKey, session.UserID)
	c.Locals(ContextSessionIDKey, session.ID)
	return c.Next()
}

func logUnauthorized(c *fiber.Ctx, reason string) {
	log.Printf("unauthorized %s %s: %s", c.Method(), c.Path(), reason)
}
