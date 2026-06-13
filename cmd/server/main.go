// Package main is the entry point for the Ainyx backend API server.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"errors"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/udayagiri/ainyx-backend/config"
	"github.com/udayagiri/ainyx-backend/internal/handler"
	"github.com/udayagiri/ainyx-backend/internal/logger"
	"github.com/udayagiri/ainyx-backend/internal/middleware"
	"github.com/udayagiri/ainyx-backend/internal/repository"
	"github.com/udayagiri/ainyx-backend/internal/routes"
	"github.com/udayagiri/ainyx-backend/internal/service"
)

func main() {
	// ── 1. Logger ──────────────────────────────────────────────────────────────
	logger.Init()
	defer logger.Sync()
	zlog := logger.Get()

	// ── 2. Config ──────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	zlog.Infow("config loaded", "port", cfg.ServerPort)

	// ── 3. Database ────────────────────────────────────────────────────────────
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		zlog.Fatalw("failed to open DB", "error", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		zlog.Fatalw("failed to ping DB", "error", err)
	}
	zlog.Info("database connection established")

	// ── 4. Dependency wiring ───────────────────────────────────────────────────
	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	// ── 5. Fiber app ───────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Ainyx Backend API",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware (order matters)
	app.Use(middleware.Recover())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	// Health-check — useful for Docker / load balancer probes.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Mount API routes
	routes.Register(app, h)

	// ── 6. Graceful shutdown ───────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		zlog.Infow("starting server", "addr", addr)
		if err := app.Listen(addr); err != nil {
			zlog.Fatalw("server error", "error", err)
		}
	}()

	<-quit
	zlog.Info("shutdown signal received")
	if err := app.Shutdown(); err != nil {
		zlog.Errorw("graceful shutdown error", "error", err)
	}
	zlog.Info("server stopped")
}

// customErrorHandler is a catch-all for unhandled Fiber errors.
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
