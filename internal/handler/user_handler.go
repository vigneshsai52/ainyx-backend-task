// Package handler wires HTTP requests to the service layer via GoFiber.
package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/udayagiri/ainyx-backend/internal/models"
	"github.com/udayagiri/ainyx-backend/internal/repository"
	"github.com/udayagiri/ainyx-backend/internal/service"
	"github.com/udayagiri/ainyx-backend/internal/logger"
)

// UserHandler holds dependencies for user-related HTTP handlers.
type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

// New creates a UserHandler.
func New(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ── POST /users ───────────────────────────────────────────────────────────────

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	log := logger.Get()

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warnw("invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: err.Error()})
	}

	resp, err := h.svc.Create(c.Context(), req)
	if err != nil {
		log.Errorw("create user error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ── GET /users/:id ────────────────────────────────────────────────────────────

// GetUser handles GET /users/:id.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid id"})
	}

	resp, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ── PUT /users/:id ────────────────────────────────────────────────────────────

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	log := logger.Get()

	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid id"})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warnw("invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: err.Error()})
	}

	resp, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ── DELETE /users/:id ─────────────────────────────────────────────────────────

// DeleteUser handles DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ── GET /users ────────────────────────────────────────────────────────────────

// ListUsers handles GET /users with optional ?page= and ?page_size= params.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)

	resp, err := h.svc.List(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseID(c *fiber.Ctx) (int32, error) {
	raw := c.Params("id")
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return int32(id), nil
}
