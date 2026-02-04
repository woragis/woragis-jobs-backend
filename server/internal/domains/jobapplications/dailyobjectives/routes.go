package dailyobjectives

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers daily objectives routes.
func RegisterRoutes(app fiber.Router, db *gorm.DB, logger *slog.Logger) {
	repo := NewGormRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service)

	// Daily objectives management
	app.Post("/daily-objectives", handler.CreateObjective)
	app.Get("/daily-objectives", handler.GetObjective)
	app.Patch("/daily-objectives", handler.UpdateObjective)

	// Daily progress tracking
	app.Get("/daily-progress/today", handler.GetTodayProgress)
	app.Get("/daily-progress/history", handler.GetHistoricalProgress)
}
