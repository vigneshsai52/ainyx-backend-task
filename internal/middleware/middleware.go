// Package middleware provides GoFiber middleware for request tracing and logging.
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/udayagiri/ainyx-backend/internal/logger"
)

const requestIDHeader = "X-Request-ID"

// RequestID injects a unique X-Request-ID header into every response.
// If the incoming request already carries the header, its value is preserved.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		// Make the ID available to downstream handlers via locals.
		c.Locals("requestID", id)
		c.Set(requestIDHeader, id)
		return c.Next()
	}
}

// Logger logs method, path, status code, and duration for every request using
// Uber Zap.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Let the chain execute.
		err := c.Next()

		duration := time.Since(start)
		requestID, _ := c.Locals("requestID").(string)

		log := logger.Get()
		log.Infow("request",
			"requestID", requestID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", duration.String(),
			"ip", c.IP(),
		)

		return err
	}
}

// Recover catches panics, logs them, and returns HTTP 500 so the server keeps
// running.
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log := logger.Get()
				log.Errorw("panic recovered", "panic", r, "path", c.Path())
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "internal server error",
				})
			}
		}()
		return c.Next()
	}
}
