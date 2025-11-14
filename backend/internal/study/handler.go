package study

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"studytracker/internal/auth"
)

// Handler exposes HTTP endpoints for study resources.
type Handler struct {
	service *Service
}

// NewHandler creates a study handler bound to a service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts study routes onto the provided router.
// All study endpoints require authentication.
func (h *Handler) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	router.Get("/subjects", requireAuth, h.listSubjects)
	router.Post("/subjects", requireAuth, h.createSubject)
	router.Put("/subjects/:id", requireAuth, h.updateSubject)
	router.Delete("/subjects/:id", requireAuth, h.deleteSubject)

	router.Get("/study-sessions", requireAuth, h.listSessions)
	router.Post("/study-sessions", requireAuth, h.createSession)
	router.Put("/study-sessions/:id", requireAuth, h.updateSession)
	router.Delete("/study-sessions/:id", requireAuth, h.deleteSession)
	router.Get("/progress/summary", requireAuth, h.handleSummary)
}

func (h *Handler) listSessions(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListSessions(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(items)
}

func (h *Handler) createSession(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var session StudySession
	if err := c.BodyParser(&session); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	created, err := h.service.CreateSession(userID, session)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingSubject), errors.Is(err, ErrInvalidTiming), errors.Is(err, ErrUnknownSubject):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) updateSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.ErrNotFound
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var session StudySession
	if err := c.BodyParser(&session); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	session.ID = id

	updated, err := h.service.UpdateSession(userID, session)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, ErrMissingSubject), errors.Is(err, ErrInvalidTiming), errors.Is(err, ErrUnknownSubject):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(updated)
}

func (h *Handler) deleteSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.ErrNotFound
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteSession(userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) handleSummary(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	summary, err := h.service.BuildSummary(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(summary)
}

// Subject handlers ------------------------------------------------------------

func (h *Handler) listSubjects(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	subjects, err := h.service.ListSubjects(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(subjects)
}

func (h *Handler) createSubject(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var subject Subject
	if err := c.BodyParser(&subject); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	created, err := h.service.CreateSubject(userID, subject)
	if err != nil {
		switch {
		case errors.Is(err, ErrSubjectNameEmpty), errors.Is(err, ErrSubjectNameExists):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) updateSubject(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.ErrNotFound
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var subject Subject
	if err := c.BodyParser(&subject); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	subject.ID = id

	updated, err := h.service.UpdateSubject(userID, subject)
	if err != nil {
		switch {
		case errors.Is(err, ErrSubjectNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, ErrSubjectNameEmpty), errors.Is(err, ErrSubjectNameExists):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(updated)
}

func (h *Handler) deleteSubject(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.ErrNotFound
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteSubject(userID, id); err != nil {
		if errors.Is(err, ErrSubjectNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func userIDFromCtx(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals(auth.ContextUserIDKey).(string)
	if !ok || userID == "" {
		return "", fiber.ErrUnauthorized
	}
	return userID, nil
}
