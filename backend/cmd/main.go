package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/rillyayidan/nowhere/backend/backend/internal/handler"
	"github.com/rillyayidan/nowhere/backend/backend/internal/service"
)

func main() {
	_ = godotenv.Load("backend/.env.local")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load(".env")

	app := fiber.New(fiber.Config{
		AppName:      "NowHere API",
		ServerHeader: "nowhere",
	})
	app.Use(logger.New())
	allowOrigins := os.Getenv("CORS_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:5173,http://127.0.0.1:5173"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	prefs := service.NewPreferenceStore()
	handlers := handler.New(service.NewContextService(), service.NewDecisionService(prefs), prefs)
	handlers.Register(app.Group("/api"))
	handlers.Register(app)
	registerFrontend(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}

func registerFrontend(app *fiber.App) {
	staticDir := os.Getenv("FRONTEND_DIST_DIR")
	if staticDir == "" {
		staticDir = filepath.Join("frontend", "dist")
	}
	if _, err := os.Stat(staticDir); err != nil {
		return
	}

	app.Static("/", staticDir)
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(staticDir, "index.html"))
	})
}
