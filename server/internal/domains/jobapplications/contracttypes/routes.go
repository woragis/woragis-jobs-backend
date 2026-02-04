package contracttypes

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers contract type routes.
func RegisterRoutes(app fiber.Router, db *gorm.DB, logger *slog.Logger) {
	repo := NewGormRepository(db)
	service := NewService(repo, logger)
	handler := NewHandler(service)

	// Public routes
	app.Get("/contract-types", handler.ListTypes)
	app.Get("/contract-types/:id", handler.GetType)
}
