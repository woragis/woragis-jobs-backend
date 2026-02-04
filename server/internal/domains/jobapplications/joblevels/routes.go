package joblevels

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers job level routes.
func RegisterRoutes(app fiber.Router, db *gorm.DB, logger *slog.Logger) {
	repo := NewGormRepository(db)
	service := NewService(repo, logger)
	handler := NewHandler(service)

	// Public routes
	app.Get("/job-levels", handler.ListLevels)
	app.Get("/job-levels/:id", handler.GetLevel)
}
