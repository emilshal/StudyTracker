package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

// Handler exposes auth-related HTTP endpoints.
type Handler struct {
	service     *Service
	cookieName  string
	frontendURL string
}

// NewHandler creates an auth handler.
func NewHandler(service *Service, frontendURL string) *Handler {
	return &Handler{
		service:     service,
		cookieName:  "session_token",
		frontendURL: frontendURL,
	}
}

// RegisterRoutes wires auth endpoints under the provided router.
// Routes that require an existing session should pass through requireAuth.
func (h *Handler) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	router.Post("/register", h.register)
	router.Post("/login", h.login)
	router.Post("/logout", h.logout)
	router.Get("/google/login", h.googleLogin)
	router.Get("/google/callback", h.googleCallback)

	router.Get("/me", requireAuth, h.me)
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(c *fiber.Ctx) error {
	var body credentials
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	result, err := h.service.Register(body.Email, body.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	h.setAuthCookie(c, result.Session)
	return c.Status(fiber.StatusCreated).JSON(result.User)
}

func (h *Handler) login(c *fiber.Ctx) error {
	var body credentials
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	result, err := h.service.Login(body.Email, body.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	h.setAuthCookie(c, result.Session)
	return c.JSON(result.User)
}

func (h *Handler) logout(c *fiber.Ctx) error {
	sessionID := c.Cookies(h.cookieName)
	if sessionID != "" {
		_ = h.service.DeleteSession(sessionID)
	}
	h.clearAuthCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) me(c *fiber.Ctx) error {
	userID, ok := c.Locals(ContextUserIDKey).(string)
	if !ok || userID == "" {
		return fiber.ErrUnauthorized
	}
	u, err := h.service.GetUserByID(userID)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	return c.JSON(u)
}

func (h *Handler) googleLogin(c *fiber.Ctx) error {
	cfg := h.service.OAuthConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusBadRequest, "google login not configured")
	}

	state, err := h.service.GenerateOAuthState()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to start oauth flow")
	}
	h.setStateCookie(c, state)

	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.JSON(fiber.Map{"url": url})
}

func (h *Handler) googleCallback(c *fiber.Ctx) error {
	cfg := h.service.OAuthConfig()
	if cfg == nil {
		return fiber.NewError(fiber.StatusBadRequest, "google login not configured")
	}

	stateFromCookie := c.Cookies("oauth_state")
	state := c.Query("state")
	if stateFromCookie == "" || state != stateFromCookie {
		return fiber.NewError(fiber.StatusBadRequest, "invalid oauth state")
	}

	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing code")
	}

	result, err := h.service.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	h.setAuthCookie(c, result.Session)
	h.clearStateCookie(c)

	redirectTarget := h.frontendURL
	if redirectTarget == "" {
		redirectTarget = "/"
	}
	return c.Redirect(redirectTarget, http.StatusTemporaryRedirect)
}

func (h *Handler) setAuthCookie(c *fiber.Ctx, session Session) {
	seconds := int(time.Until(session.ExpiresAt).Seconds())
	if seconds <= 0 {
		seconds = int(24 * time.Hour / time.Second)
	}
	c.Cookie(&fiber.Cookie{
		Name:     h.cookieName,
		Value:    session.ID,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   strings.HasPrefix(h.frontendURL, "https://"),
		MaxAge:   seconds,
	})
}

func (h *Handler) clearAuthCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		MaxAge:   -1,
	})
}

func (h *Handler) setStateCookie(c *fiber.Ctx, state string) {
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		MaxAge:   300,
	})
}

func (h *Handler) clearStateCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		MaxAge:   -1,
	})
}

// GoogleCallbackHandler exposes the callback handler for alternate routes.
func (h *Handler) GoogleCallbackHandler() fiber.Handler {
	return h.googleCallback
}
