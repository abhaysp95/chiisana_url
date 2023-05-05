package main

import (
	"log"
	"os"

	"github.com/abhaysp95/chiisana_url/api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	// loading up .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// setting up router
	app := fiber.New()
	app.Use(logger.New())  // middleware for logging

	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
