// Package routes registers all application routes on a Fiber app instance.
package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/udayagiri/ainyx-backend/internal/handler"
)

// Register mounts all routes onto the provided Fiber app.
func Register(app *fiber.App, h *handler.UserHandler) {
	users := app.Group("/users")
	{
		users.Post("/", h.CreateUser)
		users.Get("/", h.ListUsers)
		users.Get("/:id", h.GetUser)
		users.Put("/:id", h.UpdateUser)
		users.Delete("/:id", h.DeleteUser)
	}
}
