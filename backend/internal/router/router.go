package router

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"studytracker/internal/auth"
	"studytracker/internal/platform/database"
	"studytracker/internal/study"
	"studytracker/internal/user"
)

// New wires the Fiber application for the API and static frontend.
func New(dsn string) (*fiber.App, error) {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	db, err := database.Open(database.Config{DSN: dsn})
	if err != nil {
		return nil, err
	}

	if err := database.ApplyMigrations(context.Background(), db); err != nil {
		db.Close()
		return nil, err
	}

	sessionRepo := study.NewSQLSessionRepository(db)
	subjectRepo := study.NewSQLSubjectRepository(db)
	userRepo := user.NewSQLRepository(db)
	sessionStore := auth.NewSQLSessionStore(db)
	sessionTTL := parseDuration(getenv("SESSION_TTL", "24h"), 24*time.Hour)

	authService := auth.NewService(userRepo, sessionStore, auth.Config{
		SessionTTL:         sessionTTL,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	})
	authHandler := auth.NewHandler(authService, getenv("FRONTEND_URL", ""))
	authMiddleware := auth.NewMiddleware(sessionStore)

	service := study.NewService(sessionRepo, subjectRepo)
	handler := study.NewHandler(service)

	publicAPI := app.Group("/api")
	authGroup := publicAPI.Group("/auth")
	authHandler.RegisterRoutes(authGroup, authMiddleware.RequireAuth)

	handler.RegisterRoutes(publicAPI, authMiddleware.RequireAuth)

	app.Get("/oauth/callback", authHandler.GoogleCallbackHandler())

	app.Hooks().OnShutdown(func() error {
		return db.Close()
	})

	frontendDir, err := filepath.Abs("../frontend")
	if err != nil {
		db.Close()
		return nil, err
	}

	app.Static("/", frontendDir, fiber.Static{
		Compress:      true,
		CacheDuration: 0,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(frontendDir, "index.html"))
	})

	app.Use(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			return fiber.ErrNotFound
		}
		return c.SendFile(filepath.Join(frontendDir, "index.html"))
	})

	return app, nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	return fallback
}
